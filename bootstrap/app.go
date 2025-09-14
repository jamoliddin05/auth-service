package bootstrap

import (
	"app/bootstrap/closers"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	srv     *http.Server
	closers []closers.Closer
}

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

func (a *App) RegisterCloser(c closers.Closer) {
	a.closers = append(a.closers, c)
}

func (a *App) Run() error {
	return a.srv.ListenAndServe()
}

// RunWithGracefulShutdown runs server and gracefully shuts down on SIGINT/SIGTERM
func (a *App) RunWithGracefulShutdown() {
	go func() {
		if err := a.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()
	log.Println("server running on", a.srv.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	for _, c := range a.closers {
		if err := c.Close(ctx); err != nil {
			log.Printf("failed to close resource: %v", err)
		}
	}

	if err := a.srv.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown gracefully: %v", err)
	}
	log.Println("server exited cleanly")
}
