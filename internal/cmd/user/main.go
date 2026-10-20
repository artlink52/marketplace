package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/artlink52/marketplace/internal/user/app"
	"github.com/artlink52/marketplace/internal/user/config"
)

func main() {
	cfg := config.MustLoad()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.GRPCServer.Run()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Printf("server error: %v", err)
	case <-stop:
		log.Println("shutting down")
	}
	application.Stop()
}
