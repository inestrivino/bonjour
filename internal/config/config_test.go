package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inestrivino/bonjour/internal/modules/weather"
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

// Test GetConfigPath under normal conditions
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

// Test GetConfigPath failure state when environment variables prevent resolving config directory
func TestGetConfigPath_Error(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")
	t.Setenv("APPDATA", "")

	_, err := GetConfigPath()
	if err == nil {
		t.Error("expected GetConfigPath() to return error when user config dir cannot be found, got nil")
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

// Test SaveConfig error paths (e.g. invalid permissions/path)
func TestSaveConfig_Errors(t *testing.T) {
	tmpDir := t.TempDir()

	// Cause MkdirAll to fail by making destination parent a file instead of a directory
	filePath := filepath.Join(tmpDir, "file_blocking_dir")
	if err := os.WriteFile(filePath, []byte("blocker"), 0644); err != nil {
		t.Fatalf("failed setup: %v", err)
	}

	invalidSavePath := filepath.Join(filePath, "subfolder", "config.json")
	err := SaveConfig(invalidSavePath, &Config{})
	if err == nil {
		t.Error("expected SaveConfig() to fail when parent path is a file, got nil")
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

// Test LoadOrRunWizard error path when config directory resolution fails
func TestLoadOrRunWizard_GetConfigPathError(t *testing.T) {
	// Point config dir to an uncreatable location (a file instead of a folder)
	tmpDir := t.TempDir()
	blockingFile := filepath.Join(tmpDir, "blocked")
	_ = os.WriteFile(blockingFile, []byte("blocker"), 0644)

	t.Setenv("XDG_CONFIG_HOME", blockingFile)
	t.Setenv("HOME", blockingFile)
	t.Setenv("APPDATA", blockingFile)

	_, err := LoadOrRunWizard()
	if err == nil {
		t.Error("expected LoadOrRunWizard() to fail when path resolution fails, got nil")
	}
}

// Test the ValidateName helper function for the User Configuration forms
func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string", "", true},
		{"whitespace string", "   ", true},
		{"valid name", "TestUser", false},
		{"valid name with spaces", "John Doe", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// Tests that the modules that have been passed as true are set as such in the config
func TestApplySelectedModules(t *testing.T) {
	cfg := &Config{}

	// Test enabling all modules
	ApplySelectedModules(cfg, []string{"weather", "quotes", "events"})
	if !cfg.Dashboard.ShowWeather || !cfg.Dashboard.ShowQuotes || !cfg.Dashboard.ShowEvents {
		t.Errorf("expected all modules to be true, got: %+v", cfg.Dashboard)
	}

	// Test enabling a subset (weather only)
	ApplySelectedModules(cfg, []string{"weather"})
	if !cfg.Dashboard.ShowWeather || cfg.Dashboard.ShowQuotes || cfg.Dashboard.ShowEvents {
		t.Errorf("expected only ShowWeather to be true, got: %+v", cfg.Dashboard)
	}

	// Test enabling none
	ApplySelectedModules(cfg, []string{})
	if cfg.Dashboard.ShowWeather || cfg.Dashboard.ShowQuotes || cfg.Dashboard.ShowEvents {
		t.Errorf("expected all modules to be false, got: %+v", cfg.Dashboard)
	}
}

// Tests that the location is correctly set in the config when resolveLocation is successfully called
func TestResolveAndApplyLocation_Success(t *testing.T) {
	cfg := &Config{}

	// Temporarily mock resolveLocation to return fake coordinates
	oldResolve := resolveLocation
	resolveLocation = func(city, country, admin string) (*weather.LocationResult, error) {
		return &weather.LocationResult{
			Latitude:  48.8566,
			Longitude: 2.3522,
		}, nil
	}
	defer func() { resolveLocation = oldResolve }() // Restore after test

	err := ResolveAndApplyLocation(cfg, "Paris", "France", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.User.City != "Paris" || cfg.User.Latitude != 48.8566 || cfg.User.Longitude != 2.3522 {
		t.Errorf("expected location to be updated, got: %+v", cfg.User)
	}
}

// Tests that the location is correctly set empty in the configuration when resolveLocation is called empty
func TestResolveAndApplyEmptyLocation_Success(t *testing.T) {
	cfg := &Config{}

	// Temporarily mock resolveLocation returning Paris coordinates
	oldResolve := resolveLocation
	resolveLocation = func(city, country, admin string) (*weather.LocationResult, error) {
		return &weather.LocationResult{
			Name:      "",
			Latitude:  0,
			Longitude: 0,
		}, nil
	}
	defer func() { resolveLocation = oldResolve }() // Restore after test

	err := ResolveAndApplyLocation(cfg, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.User.City != "" || cfg.User.Latitude != 0 || cfg.User.Longitude != 0 {
		t.Errorf("expected location to be updated to Paris, got: %+v", cfg.User)
	}
}

// Tests that the location in the configuration is correctly set to nothing when the input place does not exist
func TestResolveAndApplyLocation_Error(t *testing.T) {
	cfg := &Config{}

	oldResolve := resolveLocation
	resolveLocation = func(city, country, admin string) (*weather.LocationResult, error) {
		return nil, fmt.Errorf("api failure")
	}
	defer func() { resolveLocation = oldResolve }()

	err := ResolveAndApplyLocation(cfg, "BadCity", "Nowhere", "")
	if err == nil {
		t.Error("expected an error when resolution fails, got nil")
	}
}
