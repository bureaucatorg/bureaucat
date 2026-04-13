package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config stores Bureaucat CLI settings on disk.
type Config struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// ConfigPath returns the config path, optionally overridden by env.
func ConfigPath() string {
	if path := strings.TrimSpace(os.Getenv("BUREAUCAT_CONFIG_PATH")); path != "" {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".bureaucat-config.json")
	}

	return filepath.Join(home, ".config", "bureaucat", "config.json")
}

// LoadConfig reads config from disk. Missing config is not an error.
func LoadConfig() (Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}

	return cfg, nil
}

// SaveConfig writes config with secure permissions.
func SaveConfig(cfg Config) (string, error) {
	path := ConfigPath()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}

	return path, nil
}

// ClearConfig removes the config file if it exists.
func ClearConfig() error {
	path := ConfigPath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove config: %w", err)
	}
	return nil
}

// GetCredentials returns base URL and token with env vars overriding config.
func GetCredentials() (string, string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", "", err
	}

	url := strings.TrimSpace(os.Getenv("BUREAUCAT_URL"))
	token := strings.TrimSpace(os.Getenv("BUREAUCAT_TOKEN"))

	if url == "" {
		url = strings.TrimSpace(cfg.URL)
	}
	if token == "" {
		token = strings.TrimSpace(cfg.Token)
	}

	if url == "" || token == "" {
		return "", "", fmt.Errorf("not authenticated: run 'bureaucat login' or set BUREAUCAT_URL and BUREAUCAT_TOKEN")
	}

	return url, token, nil
}
