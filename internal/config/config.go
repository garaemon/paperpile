package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	configDirName  = "paperpile"
	configFileName = "config"
	configFileType = "yaml"
)

// Load initializes viper and reads the config file if it exists.
func Load() error {
	configDir, err := resolveConfigDir()
	if err != nil {
		return fmt.Errorf("failed to resolve config directory: %w", err)
	}

	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)
	viper.AddConfigPath(configDir)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil
		}
		return fmt.Errorf("failed to read config: %w", err)
	}
	return nil
}

// SaveSession persists the plack_session value to the config file.
func SaveSession(session string) error {
	configDir, err := resolveConfigDir()
	if err != nil {
		return fmt.Errorf("failed to resolve config directory: %w", err)
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.Set("plack_session", session)

	configPath := filepath.Join(configDir, configFileName+"."+configFileType)
	return viper.WriteConfigAs(configPath)
}

// GetSession returns the stored plack_session value.
func GetSession() string {
	return viper.GetString("plack_session")
}

func resolveConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", configDirName), nil
}
