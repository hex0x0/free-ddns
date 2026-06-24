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

// Load reads the configuration from path and unmarshals YAML into Config.
func Load(path string) error {
	if Config != nil {
		return nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file failed, path: %s err: %+v", path, err)
	}

	if err = yaml.Unmarshal(b, &Config); err != nil {
		return fmt.Errorf("parse YAML config failed, path: %s err: %+v", path, err)
	}

	return nil
}
