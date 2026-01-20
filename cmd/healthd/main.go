package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"services-health-check/internal/app"
)

func main() {
	configPath := flag.String("config", "configs/example.yaml", "config file path")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, *configPath); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
