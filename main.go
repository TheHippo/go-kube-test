package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogchi "github.com/samber/slog-chi"
)

const defaultPort = 8080

func main() {
	logger := slog.Default()
	r := chi.NewRouter()

	requestLogger := logger.With("type", "request")

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(slogchi.New(requestLogger))
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello World")) //nolint: errcheck
	})

	r.Get("/{name}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(fmt.Sprintf("Hello %s", chi.URLParam(r, "name")))) //nolint: errcheck
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", defaultPort),
		Handler: r,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		if err := server.Close(); err != nil {
			logger.LogAttrs(context.Background(), slog.LevelWarn, "Server stopped", slog.String("err", err.Error()))
		}
	}()

	logger.LogAttrs(context.Background(), slog.LevelInfo, "Starting server", slog.Int("port", defaultPort))
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.LogAttrs(context.Background(), slog.LevelError, "Could not start server", slog.String("err", err.Error()))
	}
}
