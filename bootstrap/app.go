package bootstrap

import (
	"app/bootstrap/closers"
	"context"
	"net/http"
	"time"
)

type App struct {
	srv     *http.Server
	closers []closers.Closer
}

// NewApp builds an App with any http.Handler (Gin, Chi, net/http mux, etc.)
func NewApp(handler http.Handler, addr string) *App {
	return &App{
		srv: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// RegisterCloser adds resources (DBs, caches, brokers) to close on shutdown.
func (a *App) RegisterCloser(c closers.Closer) {
	a.closers = append(a.closers, c)
}

// Run starts the HTTP server (blocking).
func (a *App) Run() error {
	return a.srv.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server and closes all resources.
func (a *App) Shutdown(ctx context.Context) error {
	// close resources first
	for _, c := range a.closers {
		if err := c.Close(ctx); err != nil {
			return err
		}
	}
	// then gracefully stop the HTTP server
	return a.srv.Shutdown(ctx)
}
