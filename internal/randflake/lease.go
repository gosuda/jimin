package randflake

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"gosuda.org/jimin/database"
)

var processUUID = uuid.Must(uuid.NewV7())

const (
	_LEASE_SAFE_DURATION    = time.Minute * 3
	_LEASE_RENEWAL_DURATION = time.Minute * 7
	_LEASE_DURATION         = time.Minute * 10
	_LEASE_RENEWAL_INTERVAL = time.Second * 10
	_LEASE_RANGE_MAX        = 1<<17 - 1
	_LEASE_SIZE             = 128
	_LEASE_DB_FETCH_SIZE    = 1024
)

type lease struct {
	node   database.RandflakeNode
	nidmap [_LEASE_SIZE / 64]uint64
	expire atomic.Int64
}

var (
	ErrResourceExhausted = errors.New("resource exhausted")
	errRetryable         = errors.New("retryable")
)

func tryLease(ctx context.Context, c *pgxpool.Conn) (*lease, error) {
	tx, err := c.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	db := database.New(c).WithTx(tx)

	var cursor int64 = 0
	var last_end int64 = -1
	var found bool
	var free_start, free_end int64

	// scan for empty slots
L:
	for cursor < _LEASE_RANGE_MAX-_LEASE_SIZE {
		ranges, err := db.GetRandflakeRangeLease(
			ctx,
			database.GetRandflakeRangeLeaseParams{
				RangeStart: cursor,
				LeaseEnd:   time.Now().Unix(),
				Limit:      _LEASE_DB_FETCH_SIZE,
			},
		)
		if err != nil && err != pgx.ErrNoRows {
			return nil, err
		}

		if err == pgx.ErrNoRows || len(ranges) == 0 {
			free_start = cursor
			free_end = cursor + _LEASE_SIZE - 1
			found = true
			break
		}

		for i := range ranges {
			if ranges[i].RangeStart > last_end+1 {
				if last_end-ranges[i].RangeStart+1 >= _LEASE_SIZE {
					free_start = last_end + 1
					free_end = ranges[i].RangeStart
					found = true
					break L
				}
			}
			cursor = ranges[i].RangeEnd + 1
		}
	}

	if !found {
		return nil, ErrResourceExhausted
	}

	if free_end-free_start+1 < _LEASE_SIZE {
		return nil, ErrResourceExhausted
	} else {
		free_end = free_start + _LEASE_SIZE - 1
	}

	now := time.Now()

	// try allocate lease
	node, err := db.TryCreateRandflakeRangeLease(
		ctx,
		database.TryCreateRandflakeRangeLeaseParams{
			RangeStart: free_start,
			RangeEnd:   free_end - 1,
			LeaseHolder: pgtype.UUID{
				Bytes: processUUID,
				Valid: true,
			},
			LeaseStart: now.Unix(),
			LeaseEnd:   now.Add(_LEASE_DURATION).Unix(),
			LeaseEnd_2: now.Unix(),
		},
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "40001" {
			return nil, errRetryable
		}

		if err == pgx.ErrNoRows {
			return nil, errRetryable
		}
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "40001" {
			return nil, errRetryable
		}
		return nil, err
	}

	lease := lease{
		node: node,
	}
	lease.expire.Store(now.Add(_LEASE_DURATION).Unix())
	return &lease, nil
}

func newLease(ctx context.Context, c *pgxpool.Conn) (*lease, error) {
	for {
		lease, err := tryLease(ctx, c)
		if err != nil {
			if err == errRetryable {
				time.Sleep(time.Millisecond * 100)
				continue
			}
			return nil, err
		}
		return lease, nil
	}
}

func (l *lease) renewLease(ctx context.Context, c *pgxpool.Conn) error {
	tx, err := c.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	db := database.New(c).WithTx(tx)
	now := time.Now()

	node, err := db.RenewRandflakeRangeLease(
		ctx,
		database.RenewRandflakeRangeLeaseParams{
			LeaseEnd:    now.Add(_LEASE_DURATION).Unix(),
			LeaseEnd_2:  now.Unix(),
			ID:          l.node.ID,
			LeaseHolder: l.node.LeaseHolder,
		},
	)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	l.node = node
	l.expire.Store(now.Add(_LEASE_DURATION).Unix())

	return nil
}
