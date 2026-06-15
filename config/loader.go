package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultPath returns the default location of the configuration file:
//
// $HOME/.config/free-ddns/config.yaml
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home dir: %w", err)
	}
	return filepath.Join(home, ".config", "free-ddns", "config.yaml"), nil
}

var Config *AppConfig

// MustLoad reads the configuration from path and unmarshals YAML into Config.
func MustLoad(path string) {
	if Config != nil {
		return
	}

	b, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("read config file %q: %w", path, err))
	}

	if err := yaml.Unmarshal(b, &Config); err != nil {
		panic(fmt.Errorf("parse YAML config %q: %w", path, err))
	}
}
