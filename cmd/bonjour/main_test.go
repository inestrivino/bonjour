package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inestrivino/bonjour/internal/config"
	"github.com/inestrivino/bonjour/internal/ui"
)

// Helper to isolate config environment and pre-seed config to bypass wizard
func setupTestConfig(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Override config paths for all operating systems
	t.Setenv("XDG_CONFIG_HOME", tmpDir) // Linux / BSD
	t.Setenv("HOME", tmpDir)            // macOS
	t.Setenv("APPDATA", tmpDir)         // Windows

	// Pre-create configuration directory and seed file
	configDir := filepath.Join(tmpDir, "bonjour")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create mock config directory: %v", err)
	}

	testConfig := &config.Config{
		User: config.UserConfig{
			Name:      "TestUser",
			City:      "Madrid",
			Latitude:  40.4168,
			Longitude: -3.7038,
		},
		Dashboard: config.DashboardConfig{
			Theme:       "charm",
			ShowWeather: true,
			ShowQuotes:  true,
			ShowEvents:  true,
		},
		Events: []config.EventConfig{
			{
				Title:        "Sample Event",
				Date:         "2026-12-25",
				WarningStart: "2026-12-20",
			},
		},
	}

	configPath := filepath.Join(configDir, "config.json")
	if err := config.SaveConfig(configPath, testConfig); err != nil {
		t.Fatalf("failed to seed mock config file: %v", err)
	}

	return tmpDir
}

// Helper to capture stdout produced during execution
func captureStdout(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// Test module visibility logic
func TestDetermineModulesToShow(t *testing.T) {
	tests := []struct {
		name        string
		opts        CLIOptions
		cfg         config.Config
		wantQuotes  bool
		wantWeather bool
		wantEvents  bool
	}{
		{
			name: "All enabled in config and no CLI flags",
			opts: CLIOptions{},
			cfg: config.Config{
				User: config.UserConfig{Latitude: 40.7128, Longitude: -74.0060},
				Dashboard: config.DashboardConfig{
					ShowQuotes:  true,
					ShowWeather: true,
					ShowEvents:  true,
				},
			},
			wantQuotes:  true,
			wantWeather: true,
			wantEvents:  true,
		},
		{
			name: "CLI flags disable all modules",
			opts: CLIOptions{
				NoQuotes:  true,
				NoWeather: true,
				NoEvents:  true,
			},
			cfg: config.Config{
				User: config.UserConfig{Latitude: 40.7128, Longitude: -74.0060},
				Dashboard: config.DashboardConfig{
					ShowQuotes:  true,
					ShowWeather: true,
					ShowEvents:  true,
				},
			},
			wantQuotes:  false,
			wantWeather: false,
			wantEvents:  false,
		},
		{
			name: "Weather disabled due to zero coordinates (0, 0)",
			opts: CLIOptions{},
			cfg: config.Config{
				User: config.UserConfig{Latitude: 0, Longitude: 0},
				Dashboard: config.DashboardConfig{
					ShowQuotes:  true,
					ShowWeather: true,
					ShowEvents:  true,
				},
			},
			wantQuotes:  true,
			wantWeather: false,
			wantEvents:  true,
		},
		{
			name: "Disabled in config with no CLI flags",
			opts: CLIOptions{},
			cfg: config.Config{
				User: config.UserConfig{Latitude: 40.7128, Longitude: -74.0060},
				Dashboard: config.DashboardConfig{
					ShowQuotes:  false,
					ShowWeather: false,
					ShowEvents:  false,
				},
			},
			wantQuotes:  false,
			wantWeather: false,
			wantEvents:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuotes, gotWeather, gotEvents := determineModulesToShow(tt.opts, &tt.cfg)

			if gotQuotes != tt.wantQuotes {
				t.Errorf("determineModulesToShow() quotes = %v, want %v", gotQuotes, tt.wantQuotes)
			}
			if gotWeather != tt.wantWeather {
				t.Errorf("determineModulesToShow() weather = %v, want %v", gotWeather, tt.wantWeather)
			}
			if gotEvents != tt.wantEvents {
				t.Errorf("determineModulesToShow() events = %v, want %v", gotEvents, tt.wantEvents)
			}
		})
	}
}

// Test greeting output formatting across time slots
func TestGreeting(t *testing.T) {
	testTheme := ui.NewTheme("charm")
	testCfg := &config.Config{
		User: config.UserConfig{Name: "TestUser"},
	}

	tests := []struct {
		name     string
		hour     int
		expected string
	}{
		{"Morning greeting", 9, "morning"},
		{"Afternoon greeting", 14, "afternoon"},
		{"Evening greeting", 20, "evening"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				greeting(testCfg, testTheme, tt.hour)
			})

			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected greeting to contain %q for hour %d, got: %s", tt.expected, tt.hour, output)
			}
			if !strings.Contains(output, "TestUser") {
				t.Errorf("expected greeting to contain 'TestUser', got: %s", output)
			}
		})
	}
}

// Test cobra command definitions and structure
func TestCobraCommandSetup(t *testing.T) {
	if rootCmd.Use != "bonjour" {
		t.Errorf("expected rootCmd.Use to be 'bonjour', got %s", rootCmd.Use)
	}

	var foundConfig, foundEvents bool
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "config" {
			foundConfig = true
		}
		if cmd.Name() == "events" {
			foundEvents = true
		}
	}

	if !foundConfig {
		t.Error("expected 'config' subcommand to be attached to rootCmd")
	}
	if !foundEvents {
		t.Error("expected 'events' subcommand to be attached to rootCmd")
	}
}

// Test flags setup
func TestCLIFlags(t *testing.T) {
	flags := []string{"noweather", "noquotes", "noevents", "mini"}

	for _, flagName := range flags {
		flag := rootCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("expected CLI flag '--%s' to be defined", flagName)
		}
	}
}

// Test application execution pipeline in mini mode
func TestRunApplication_MiniMode(t *testing.T) {
	_ = setupTestConfig(t) // Seeds config file so wizard is bypassed

	opts := CLIOptions{
		Mini:      true,
		NoWeather: true,
		NoQuotes:  true,
		NoEvents:  true,
	}

	output := captureStdout(func() {
		_ = runApplication(opts)
	})

	// Mini mode greeting contains the user's name but omits full card structures
	if !strings.Contains(output, "TestUser") {
		t.Errorf("expected output to contain user name 'TestUser', got: %s", output)
	}
}

// Test application execution pipeline in full mode
func TestRunApplication_FullMode(t *testing.T) {
	_ = setupTestConfig(t) // Seeds config file so wizard is bypassed

	// Configure the root command's arguments to match basic options
	rootCmd.SetArgs([]string{
		"--noweather",
		"--noquotes",
		"--noevents",
	})

	var output string
	output = captureStdout(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	if !strings.Contains(output, "TestUser") {
		t.Errorf("expected output to contain user name 'TestUser', got: %s", output)
	}
}

// Test configuration execution pipeline
func TestConfigCommand_Execution(t *testing.T) {
	_ = setupTestConfig(t) //Seeds config file so wizard is bypassed

	// Temporarily replace the wizard with a mock that does nothing and returns no error
	oldWizard := runConfigWizard
	runConfigWizard = func(theme *ui.Theme) error {
		return nil
	}
	defer func() { runConfigWizard = oldWizard }() // Restore it after the test

	// CLI arguments should point to the config command
	rootCmd.SetArgs([]string{"config"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("expected no error from 'config' command execution, got %v", err)
	}
}

// Test events configuration execution pipeline
func TestEventsCommand_Execution(t *testing.T) {
	_ = setupTestConfig(t) //Seeds config file so wizard is bypassed

	// Temporarily replace the wizard with a mock that does nothing and returns no error
	oldWizard := runEventsWizard
	runEventsWizard = func(theme *ui.Theme) error {
		return nil
	}
	defer func() { runEventsWizard = oldWizard }() // Restore it after the test

	// CLI arguments should point to the events command
	rootCmd.SetArgs([]string{"events"})

	// Execute the command
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("expected no error from 'events' command execution, got %v", err)
	}
}

// TestRenderComponents covers all layout configurations and condition branches inside renderComponents
func TestRenderComponents(t *testing.T) {
	_ = setupTestConfig(t)
	testTheme := ui.NewTheme("charm")
	testCfg := &config.Config{
		User: config.UserConfig{
			Name:      "TestUser",
			City:      "Madrid",
			Latitude:  40.4168,
			Longitude: -3.7038,
		},
		Dashboard: config.DashboardConfig{
			Theme:       "charm",
			ShowWeather: true,
			ShowQuotes:  true,
			ShowEvents:  true,
		},
	}

	tests := []struct {
		name        string
		showQuotes  bool
		showWeather bool
		showEvents  bool
		mini        bool
	}{
		{
			name:        "Full rendering - All modules active",
			showQuotes:  true,
			showWeather: true,
			showEvents:  true,
			mini:        false,
		},
		{
			name:        "Mini rendering - All modules active",
			showQuotes:  true,
			showWeather: true,
			showEvents:  true,
			mini:        true,
		},
		{
			name:        "Weather only - no quotes or events",
			showQuotes:  false,
			showWeather: true,
			showEvents:  false,
			mini:        false,
		},
		{
			name:        "Events only - no quotes or weather",
			showQuotes:  false,
			showWeather: false,
			showEvents:  true,
			mini:        false,
		},
		{
			name:        "Quotes only - no weather or events",
			showQuotes:  true,
			showWeather: false,
			showEvents:  false,
			mini:        false,
		},
		{
			name:        "No modules active",
			showQuotes:  false,
			showWeather: false,
			showEvents:  false,
			mini:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				err := renderComponents(tt.showQuotes, tt.showWeather, tt.showEvents, tt.mini, testTheme, testCfg)
				if err != nil {
					t.Fatalf("unexpected error rendering components: %v", err)
				}
			})

			// Ensure basic greeting executes across all variants
			if !strings.Contains(output, "TestUser") {
				t.Errorf("expected output to contain 'TestUser', got: %s", output)
			}
		})
	}
}
