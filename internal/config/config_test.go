package config

import (
	"os"
	"path/filepath"
	"testing"
)

//TODO: ADD TESTS THAT ARE SUPPOSED TO GO WRONG (NO NAME, LOCATION DOES NOT EXIST, ETC)

// TestSaceAndLoadConfig checks whether the result of writing a Config type object as a JSON and retrieving it is correct
func TestSaveAndLoadConfig(t *testing.T) {
	// Temporary directory isolated from system configuration
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	// Sample mock configuration
	originalCfg := &Config{
		User: UserConfig{
			Name:      "TestUser",
			Latitude:  40.4165,
			Longitude: -3.70256,
		},
		Dashboard: DashboardConfig{
			ShowWeather: true,
			ShowQuotes:  false,
			ShowRSS:     true,
		},
	}

	// TEST: saveConfig
	if err := SaveConfig(configPath, originalCfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify the file was written into the disk
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Expected config file to exist at %s, but it was not created", configPath)
	}

	// TEST: loadConfig
	loadedCfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	// Verify the loaded data matches the original
	if loadedCfg.User.Name != originalCfg.User.Name {
		t.Errorf("Expected user name %q, got %q", originalCfg.User.Name, loadedCfg.User.Name)
	}
	if loadedCfg.Dashboard.ShowWeather != originalCfg.Dashboard.ShowWeather {
		t.Errorf("Expected ShowWeather %v, got %v", originalCfg.Dashboard.ShowWeather, loadedCfg.Dashboard.ShowWeather)
	}
	if loadedCfg.Dashboard.ShowQuotes != originalCfg.Dashboard.ShowQuotes {
		t.Errorf("Expected ShowQuotes %v, got %v", originalCfg.Dashboard.ShowQuotes, loadedCfg.Dashboard.ShowQuotes)
	}
}
