package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/projectsvc/internal/http/server"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	srv, err := server.New(server.Args{
		Addr:        ":8080",
		CORSOrigins: os.Getenv("CORS_ORIGINS"),
		Deps:        newDeps(),
		Log:         log,
	})
	if err != nil {
		log.Error("server setup failed", "error", err)
		os.Exit(1)
	}

	go func() {
		if err := srv.Start(); err != nil {
			log.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown failed", "error", err)
	}
}
