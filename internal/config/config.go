package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey     string `yaml:"api_key"`
	E2EEnabled bool   `yaml:"e2e_enabled"`
	E2EKey     string `yaml:"e2e_key,omitempty"`

	Notifications NotificationConfig `yaml:"notifications"`
	GUI           GUIConfig          `yaml:"gui"`
	Autostart     bool               `yaml:"autostart"`
}

type NotificationConfig struct {
	Enabled     bool     `yaml:"enabled"`
	ShowMirrors bool     `yaml:"show_mirrors"`
	ShowSMS     bool     `yaml:"show_sms"`
	ShowCalls   bool     `yaml:"show_calls"`
	Filters     []string `yaml:"filters,omitempty"`
}

type GUIConfig struct {
	ShowTrayIcon   bool `yaml:"show_tray_icon"`
	StartMinimized bool `yaml:"start_minimized"`
}

func Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config or create default
	cfg := &Config{
		Notifications: NotificationConfig{
			Enabled:     true,
			ShowMirrors: true,
			ShowSMS:     true,
			ShowCalls:   true,
		},
		GUI: GUIConfig{
			ShowTrayIcon:   true,
			StartMinimized: false,
		},
		Autostart: false,
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file
		return cfg, cfg.Save(configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

func (c *Config) Save(configPath string) error {
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func getDefaultConfigPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, _ := os.UserHomeDir()
		configHome = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configHome, "pushbulleter", "config.yaml")
}
