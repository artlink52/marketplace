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

	go application.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	application.Stop()
}
