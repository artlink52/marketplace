package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/artlink52/marketplace/internal/order/app"
	"github.com/artlink52/marketplace/internal/order/config"
)

func main() {
	cfg := config.MustLoad()
	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.GrpcApp.Run()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Printf("server error: %v", err)
	case <-sigs:
		log.Println("shutting down")
	}
	application.Stop()
}
