package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pushbulleter/internal/app"
	"pushbulleter/internal/config"
)

func main() {
	var (
		configPath = flag.String("config", "", "Path to config file (default: XDG_CONFIG_HOME/pushbulleter/config.yaml)")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create application
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Run application in GUI mode
	errChan := make(chan error, 1)
	go func() {
		errChan <- application.RunGUI(ctx)
	}()

	select {
	case <-ctx.Done():
		// Signal received, stop the app
		application.Stop()
		return
	case err := <-errChan:
		if err != nil {
			log.Fatalf("Application error: %v", err)
		}
	}
}
