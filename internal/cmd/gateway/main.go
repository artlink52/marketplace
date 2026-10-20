package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/artlink52/marketplace/internal/gateway/app"
	"github.com/artlink52/marketplace/internal/gateway/config"
)

func main() {
	cfg := config.MustLoad()
	ctx := context.Background()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.Run()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-errCh:
	case <-stop:
	}

	if err := application.Stop(ctx); err != nil {
		log.Fatal(err)
	}
}
