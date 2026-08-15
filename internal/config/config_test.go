package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to isolate user config directories cross-platform
func setupTestEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	t.Setenv("XDG_CONFIG_HOME", tmpDir) // Linux / BSD
	t.Setenv("HOME", tmpDir)            // macOS
	t.Setenv("APPDATA", tmpDir)         // Windows

	return tmpDir
}

// Test GetConfigPath under normal and overridden conditions
func TestGetConfigPath(t *testing.T) {
	tmpDir := setupTestEnv(t)

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() returned unexpected error: %v", err)
	}

	if !strings.HasPrefix(path, tmpDir) {
		t.Errorf("expected path to be inside temp dir %s, got: %s", tmpDir, path)
	}

	if !strings.HasSuffix(path, filepath.Join("bonjour", "config.json")) {
		t.Errorf("expected path to end with 'bonjour/config.json', got: %s", path)
	}
}

// Test SaveConfig & LoadConfig Serialization and Deserialization
func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := setupTestEnv(t)
	targetPath := filepath.Join(tmpDir, "bonjour", "config.json")

	originalCfg := &Config{
		User: UserConfig{
			Name:      "Jane Doe",
			Latitude:  48.8566,
			Longitude: 2.3522,
			City:      "Paris",
		},
		Dashboard: DashboardConfig{
			Theme:       "dracula",
			ShowWeather: true,
			ShowQuotes:  false,
			ShowEvents:  true,
		},
		Events: []EventConfig{
			{
				Title:        "Meeting",
				Date:         "2026-09-01",
				WarningStart: "2026-08-25",
			},
		},
	}

	// Test SaveConfig
	err := SaveConfig(targetPath, originalCfg)
	if err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Verify file was written
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Fatal("SaveConfig() did not create the file on disk")
	}

	// Test LoadConfig
	loadedCfg, err := LoadConfig(targetPath)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Validate fields
	if loadedCfg.User.Name != originalCfg.User.Name {
		t.Errorf("User.Name = %s, want %s", loadedCfg.User.Name, originalCfg.User.Name)
	}
	if loadedCfg.Dashboard.Theme != originalCfg.Dashboard.Theme {
		t.Errorf("Dashboard.Theme = %s, want %s", loadedCfg.Dashboard.Theme, originalCfg.Dashboard.Theme)
	}
	if len(loadedCfg.Events) != 1 || loadedCfg.Events[0].Title != "Meeting" {
		t.Errorf("Events mismatch: %+v", loadedCfg.Events)
	}
}

// Test LoadConfig Edge Cases (Invalid Path & Malformed JSON)
func TestLoadConfig_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	// Non-existent file error
	_, err := LoadConfig(filepath.Join(tmpDir, "does_not_exist.json"))
	if err == nil {
		t.Error("LoadConfig() expected error for non-existent file, got nil")
	}

	// Malformed JSON error
	invalidJSONPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidJSONPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("failed to create dummy invalid JSON file: %v", err)
	}

	_, err = LoadConfig(invalidJSONPath)
	if err == nil {
		t.Error("LoadConfig() expected unmarshal error for malformed JSON, got nil")
	}
}

// Test LoadOrRunWizard when valid configuration already exists
func TestLoadOrRunWizard_ExistingConfig(t *testing.T) {
	_ = setupTestEnv(t)

	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() failed: %v", err)
	}

	expectedCfg := &Config{
		User: UserConfig{Name: "Existing User"},
		Dashboard: DashboardConfig{
			Theme: "catppuccin",
		},
	}

	// Pre-create existing configuration file
	if err := SaveConfig(configPath, expectedCfg); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Call LoadOrRunWizard; should directly load without running wizard
	loadedCfg, err := LoadOrRunWizard()
	if err != nil {
		t.Fatalf("LoadOrRunWizard() returned error: %v", err)
	}

	if loadedCfg.User.Name != "Existing User" {
		t.Errorf("expected user name 'Existing User', got '%s'", loadedCfg.User.Name)
	}
}
