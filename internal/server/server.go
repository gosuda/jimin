package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/ksuid"
)

type ServerStatus int32

const (
	ServerStatusStopped ServerStatus = iota
	ServerStatusStopping
	ServerStatusStarting
	ServerStatusRunning
)

var (
	ErrAlreadyRunning  = errors.New("server is already running")
	ErrInvalidListener = errors.New("invalid listener")
)

type Server struct {
	mu     sync.Mutex
	stop   chan struct{}
	status atomic.Int32

	serverID ksuid.KSUID
	ln       net.Listener
	srv      http.Server

	errsMu sync.Mutex
	errs   []error
}

func New() *Server {
	g := &Server{
		stop:     nil,
		serverID: ksuid.New(),
		srv: http.Server{
			IdleTimeout: time.Second * 30,
		},
	}

	return g
}

func (g *Server) setState(status ServerStatus) {
	g.status.Store(int32(status))
}

func (g *Server) aError(err error) {
	if err != nil {
		g.errsMu.Lock()
		g.errs = append(g.errs, err)
		g.errsMu.Unlock()
	}
}

func (g *Server) clearErrs() {
	g.errsMu.Lock()
	g.errs = nil
	g.errsMu.Unlock()
}

func (g *Server) State() ServerStatus {
	return ServerStatus(g.status.Load())
}

func (g *Server) Start(ln net.Listener) error {
	if g.State() != ServerStatusStopped {
		return ErrAlreadyRunning
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State() != ServerStatusStopped {
		return ErrAlreadyRunning
	}

	if g.ln == nil {
		return ErrInvalidListener
	}

	g.clearErrs()
	g.ln = ln
	g.stop = make(chan struct{})
	g.setState(ServerStatusStarting)

	go func() {
		defer g.doStop()
		if err := g.srv.Serve(g.ln); err != nil && err != http.ErrServerClosed {
			g.aError(err)
		}
	}()

	return nil
}

func (g *Server) doStop() error {
	if g.State() != ServerStatusRunning {
		return nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State() != ServerStatusRunning {
		return nil
	}
	g.setState(ServerStatusStopping)

	close(g.stop)

	err := g.srv.Shutdown(context.Background())
	g.aError(err)

	g.stop = nil
	g.ln = nil

	return nil
}
