package main

import (
	"context"
	"flag"
	"fmt"
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
		daemon     = flag.Bool("daemon", false, "Run as daemon without GUI")
		version    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *version {
		fmt.Println("Pushbullet Client v1.0.0")
		return
	}

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

	// Run application
	if *daemon {
		err = application.RunDaemon(ctx)
	} else {
		err = application.RunGUI(ctx)
	}

	if err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
