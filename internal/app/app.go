package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"pushbulleter/internal/config"
	"pushbulleter/internal/notifications"
	"pushbulleter/internal/pushbullet"
	"pushbulleter/internal/tray"
)

type App struct {
	config       *config.Config
	client       *pushbullet.Client
	notifManager *notifications.Manager
	trayManager  *tray.TrayManager
}

func New(cfg *config.Config) (*App, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required. Please set it in the config file")
	}

	var e2eKey string
	if cfg.E2EEnabled {
		e2eKey = cfg.E2EKey
	}

	client := pushbullet.NewClient(cfg.APIKey, e2eKey)

	notifManager := notifications.NewManager(
		cfg.Notifications.Enabled,
		cfg.Notifications.ShowMirrors,
		cfg.Notifications.ShowSMS,
		cfg.Notifications.ShowCalls,
		cfg.Notifications.Filters,
	)

	app := &App{
		config:       cfg,
		client:       client,
		notifManager: notifManager,
		trayManager:  tray.NewTrayManager(),
	}

	return app, nil
}

func (a *App) RunGUI(ctx context.Context) error {
	log.Println("Starting pushbulleter...")

	// Test API connection
	if err := a.testConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to Pushbullet API: %w", err)
	}

	// Setup autostart if enabled
	if a.config.Autostart {
		if err := a.setupAutostart(); err != nil {
			log.Printf("Failed to setup autostart: %v", err)
		}
	}

	// Start stream connection in background
	go func() {
		if err := a.client.ConnectStream(ctx, a.handleStreamMessage); err != nil {
			log.Printf("Stream connection ended: %v", err)
		}
	}()

	// Run tray (this blocks)
	a.trayManager.Run(
		func() {
			log.Println("Tray icon ready")
		},
		func() {
			log.Println("Tray icon exiting")
		},
	)

	return nil
}

func (a *App) Stop() {
	if a.trayManager != nil {
		a.trayManager.Stop()
	}
}

func (a *App) testConnection(ctx context.Context) error {
	user, err := a.client.GetUser(ctx)
	if err != nil {
		return err
	}

	if email, ok := user["email"].(string); ok {
		log.Printf("Connected as: %s", email)
	} else {
		log.Println("Connected to Pushbullet API")
	}

	// Update E2E encryption with user iden if available
	if a.config.E2EEnabled && a.config.E2EKey != "" {
		if userIden, ok := user["iden"].(string); ok {
			a.client.UpdateE2EWithUserIden(a.config.E2EKey, userIden)
			log.Println("Updated E2E encryption with user iden")
		}
	}

	return nil
}

func (a *App) handleStreamMessage(msg *pushbullet.StreamMessage) {
	// Add to events window
	HandleEvent(msg)

	// Handle push notifications
	if msg.Type == "push" && len(msg.Push) > 0 {
		var push pushbullet.Push
		if err := json.Unmarshal(msg.Push, &push); err != nil {
			log.Printf("Failed to unmarshal push: %v", err)
			return
		}

		a.notifManager.HandlePush(&push)
	}
}

func (a *App) setupAutostart() error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Create XDG autostart directory
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, _ := os.UserHomeDir()
		configHome = filepath.Join(homeDir, ".config")
	}

	autostartDir := filepath.Join(configHome, "autostart")
	if err := os.MkdirAll(autostartDir, 0755); err != nil {
		return err
	}

	// Create desktop entry
	desktopEntry := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=Pushbulleter
Comment=Pushbullet desktop client
Exec=%s
Icon=pushbullet
StartupNotify=false
NoDisplay=true
Hidden=false
X-GNOME-Autostart-enabled=true
`, execPath)

	desktopFile := filepath.Join(autostartDir, "pushbulleter.desktop")
	return os.WriteFile(desktopFile, []byte(desktopEntry), 0644)
}
