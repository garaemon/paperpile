package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestGetSession_emptyByDefault(t *testing.T) {
	viper.Reset()
	session := GetSession()
	if session != "" {
		t.Errorf("GetSession() = %q, want empty string", session)
	}
}

func TestSaveAndGetSession(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "paperpile")

	// Override HOME so resolveConfigDir points to our temp dir.
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Reconfigure viper to use the temp config directory.
	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)
	viper.AddConfigPath(configDir)

	err := SaveSession("test-session-123")
	if err != nil {
		t.Fatalf("SaveSession() error: %v", err)
	}

	got := GetSession()
	if got != "test-session-123" {
		t.Errorf("GetSession() = %q, want %q", got, "test-session-123")
	}
}

func TestLoad_noConfigFile(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
}

func TestLoad_withExistingConfigFile(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "paperpile")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := []byte("plack_session: existing-session-456\n")
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), configContent, 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	session := GetSession()
	if session != "existing-session-456" {
		t.Errorf("GetSession() = %q, want %q", session, "existing-session-456")
	}
}

func TestSaveSession_createsConfigDir(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	configDir := filepath.Join(tmpDir, ".config", "paperpile")
	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)
	viper.AddConfigPath(configDir)

	err := SaveSession("new-session-789")
	if err != nil {
		t.Fatalf("SaveSession() error: %v", err)
	}

	// Verify config directory was created
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("config directory should have been created")
	}

	// Verify config file exists
	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file should have been created")
	}

	got := GetSession()
	if got != "new-session-789" {
		t.Errorf("GetSession() = %q, want %q", got, "new-session-789")
	}
}
