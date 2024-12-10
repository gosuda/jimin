package randflake

import (
	"context"
	"encoding/binary"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"gosuda.org/jimin/internal/randflake/sparx64"
)

const (
	_RANDFLAKE_EPOCH_OFFSET  = 1730000000
	_RANDFLAKE_MAX_COUNTER   = 1<<17 - 1
	_RANDFLAKE_MAX_TIMESTAMP = 1<<30 - 1 // lifetime of 34 years
)

var (
	ErrRandflakeDead = errors.New("randflake: the randflake is dead after 34 years of use")
	ErrInvalidSecret = errors.New("randflake: invalid secret, secret must be 16 bytes long")
	ErrClosed        = errors.New("randflake: closed")
)

type RandFlake struct {
	mu     sync.Mutex
	db     *pgxpool.Pool
	leases []*lease

	poolMutex  sync.Mutex
	sourcePool []*source

	box  *sparx64.Sparx64
	stop chan struct{}
}

func NewRandFlake(db *pgxpool.Pool, secret []byte) (*RandFlake, error) {
	if len(secret) != 16 {
		return nil, ErrInvalidSecret
	}

	randflake := &RandFlake{
		db:  db,
		box: sparx64.NewSparx64(secret),
	}

	return randflake, nil
}

func (g *RandFlake) worker() {
	ticker := time.NewTicker(_LEASE_RENEWAL_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-g.stop:
			return
		case <-ticker.C:
			{
				g.mu.Lock()

				var dropList []*lease
				for _, lease := range g.leases {
					if lease.expire.Load() < time.Now().Unix()+int64(_LEASE_SAFE_DURATION/time.Second) {
						dropList = append(dropList, lease)
						continue
					}
					ctx := context.Background()

					c, err := g.db.Acquire(ctx)
					if err != nil {
						log.Error().Err(err).Msg("randflake: failed to acquire connection")
						continue
					}

					if lease.expire.Load() < time.Now().Unix()+int64(_LEASE_RENEWAL_DURATION/time.Second) {
						lease.renewLease(ctx, c)
					}
					c.Release()
				}

				for _, lease := range dropList {
					for i, l := range g.leases {
						if l == lease {
							g.leases = append(g.leases[:i], g.leases[i+1:]...)
							break
						}
					}
				}

				g.mu.Unlock()
			}
		}
	}
}

func (g *RandFlake) NewGenerator(ctx context.Context) (*Generator, error) {
	gen := &Generator{
		randflake: g,
	}

	_, err := gen.Generate(ctx)
	if err != nil {
		return nil, err
	}

	return gen, nil
}

func (g *RandFlake) getSource(t int64) (*source, error) {
	g.poolMutex.Lock()
	defer g.poolMutex.Unlock()
	for {
		if len(g.sourcePool) == 0 {
			return nil, ErrResourceExhausted
		}
		source := g.sourcePool[len(g.sourcePool)-1]
		g.sourcePool = g.sourcePool[:len(g.sourcePool)-1]
		if source.lease.expire.Load() < t+int64(_LEASE_SAFE_DURATION/time.Second) {
			continue
		}
		return source, nil
	}
}

func (g *RandFlake) putSource(s *source, t int64) {
	if s.lease.expire.Load() < t+int64(_LEASE_SAFE_DURATION/time.Second) {
		return
	}

	g.poolMutex.Lock()
	g.sourcePool = append(g.sourcePool, s)
	g.poolMutex.Unlock()
}

type source struct {
	randflake     *RandFlake
	nodeID        int64
	lease         *lease
	last_rollover atomic.Int64
	counter       atomic.Int64
}

func (g *RandFlake) newSource(ctx context.Context) (*source, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	var free int64 = -1
	var lease *lease

L:
	for i := range g.leases {
		for j := range g.leases[i].nidmap {
			if g.leases[i].nidmap[j] == 1 {
				continue
			}

			if g.leases[i].expire.Load() < time.Now().Unix()+int64(_LEASE_SAFE_DURATION/time.Second) {
				continue
			}

			for k := 0; k < 64; k++ {
				if (g.leases[i].nidmap[j] & (1 << k)) == 0 {
					candidate := g.leases[i].node.RangeStart + int64(j*64+k)
					if candidate > g.leases[i].node.RangeEnd {
						continue
					}

					free = candidate
					lease = g.leases[i]
					break L
				}
			}
		}
	}

	if free == -1 || lease == nil {
		c, err := g.db.Acquire(ctx)
		if err != nil {
			return nil, err
		}
		defer c.Release()

		lease, err := newLease(ctx, c)
		if err != nil {
			return nil, err
		}

		if lease == nil {
			log.Error().Msg("randflake: newLease() returned nil (unreachable)")
			return nil, ErrResourceExhausted
		}

		g.leases = append(g.leases, lease)

		goto L
	}

	source := source{
		randflake: g,
		nodeID:    free,
		lease:     lease,
	}

	return &source, nil
}

func (s *source) nextID() (int64, error) {
	for {
		t := time.Now().Unix()

		// check if lease is in safe duration
		if s.lease.expire.Load() < t+int64(_LEASE_SAFE_DURATION/time.Second) {
			return 0, ErrResourceExhausted
		}

		// assume 64 bit counter can't overflow in one second.
		ctr := s.counter.Add(1)
		if ctr >= _RANDFLAKE_MAX_COUNTER {
			last_rollover := s.last_rollover.Load()
			if t < last_rollover {
				// time went backwards
				// should never happen but just in case
				log.Error().Msg("randflake: timestamp consistency violation (unreachable)")
				time.Sleep(time.Millisecond * 100)
				continue
			}

			if t == last_rollover {
				// too many IDs generated in the same second
				return 0, ErrResourceExhausted
			}

			// t > last_rollover
			// do rollover
			if s.last_rollover.CompareAndSwap(last_rollover, t) {
				s.counter.Store(0)
			}
			continue
		}

		// 30 bits for timestamp
		// 17 bits for node ID
		// 17 bits for counter
		timestamp := int64(t-_RANDFLAKE_EPOCH_OFFSET) & (1<<30 - 1)
		node_id := s.nodeID & (1<<17 - 1)
		counter := ctr & (1<<17 - 1)
		return (timestamp << 34) | (node_id << 17) | counter, nil
	}
}

type Generator struct {
	randflake *RandFlake
	sources   atomic.Pointer[[]*source]
	ref       atomic.Int64

	updateMu sync.Mutex
	closed   bool
}

func (g *Generator) nextID(ctx context.Context) (int64, error) {
	g.ref.Add(1)
	defer g.ref.Add(-1)

	for {
		sources := g.sources.Load()
		if sources == nil {
			{
				g.updateMu.Lock()

				if g.closed {
					g.updateMu.Unlock()
					return 0, ErrClosed
				}

				t := time.Now().Unix()

				// fast path
				s, err := g.randflake.getSource(t)
				if err != nil && err != ErrResourceExhausted {
					g.updateMu.Unlock()
					return 0, err
				}

				// slow path
				if err == ErrResourceExhausted {
					s, err = g.randflake.newSource(ctx)
					if err != nil {
						g.updateMu.Unlock()
						return 0, err
					}
				}

				sources = &[](*source){s}
				if !g.sources.CompareAndSwap(nil, sources) {
					// unreachable
					log.Error().Msg("randflake: sources pointer changed (unreachable)")
					g.updateMu.Unlock()
					continue
				}

				g.updateMu.Unlock()
			}
		}

		for i := range *sources {
			id, err := (*sources)[i].nextID()
			if err == nil {
				return id, nil
			}

			if err == ErrRandflakeDead {
				return 0, ErrRandflakeDead
			}
		}

		// all sources are exhausted
		// try to get new sources

		{
			g.updateMu.Lock()

			if g.closed {
				g.updateMu.Unlock()
				return 0, ErrClosed
			}

			if g.sources.Load() != sources {
				g.updateMu.Unlock()
				continue
			}

			t := time.Now().Unix()

			// fast path
			s, err := g.randflake.getSource(t)
			if err != nil && err != ErrResourceExhausted {
				g.updateMu.Unlock()
				return 0, err
			}

			// slow path
			if err == ErrResourceExhausted {
				s, err = g.randflake.newSource(ctx)
				if err != nil {
					g.updateMu.Unlock()
					return 0, err
				}
			}

			newSources := make([]*source, len(*sources)+1)
			copy(newSources, *sources)
			newSources[len(*sources)] = s

			if !g.sources.CompareAndSwap(sources, &newSources) {
				// unreachable
				log.Error().Msg("randflake: sources pointer changed (unreachable)")
				g.updateMu.Unlock()
				continue
			}

			g.updateMu.Unlock()
		}
	}
}

func (g *Generator) Close() {
	g.updateMu.Lock()
	defer g.updateMu.Unlock()
	g.closed = true

	sources := g.sources.Load()
	if sources == nil {
		return
	}

	for g.ref.Load() > 0 {
		runtime.Gosched()
	}

	for i := range *sources {
		g.randflake.putSource((*sources)[i], time.Now().Unix())
	}
	g.sources.Store(nil)
}

// GenerateUnencrypted generates a unique ID without encrypting it.
// This is useful for cases where the ID does not need to be encrypted.
// The generated ID is returned as an int64.
func (g *Generator) GenerateUnencrypted(ctx context.Context) (int64, error) {
	id, err := g.nextID(ctx)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Generate generates a unique, encrypted ID.
func (g *Generator) Generate(ctx context.Context) (int64, error) {
	id, err := g.nextID(ctx)
	if err != nil {
		return 0, err
	}

	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(id))
	g.randflake.box.Encrypt(b[:], b[:])
	return int64(binary.BigEndian.Uint64(b[:])), nil
}

// GeneratePair generates unique, a encrypted ID and a unencrypted ID.
func (g *Generator) GeneratePair(ctx context.Context) (unencrypted, encrypted int64, err error) {
	id, err := g.nextID(ctx)
	if err != nil {
		return 0, 0, err
	}

	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(id))
	g.randflake.box.Encrypt(b[:], b[:])

	return id, int64(binary.BigEndian.Uint64(b[:])), nil
}
