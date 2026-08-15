package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// Test Module Visibility Logic (All Branch Combinations)
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

// Test Greeting output formatting across time slots
func TestGreeting(t *testing.T) {
	testTheme := ui.NewTheme("charm")
	testCfg := &config.Config{
		User: config.UserConfig{Name: "Alice"},
	}

	output := captureStdout(func() {
		greeting(testCfg, testTheme)
	})

	if !strings.Contains(output, "Alice") {
		t.Errorf("expected greeting output to contain 'Alice', got: %s", output)
	}

	hour := time.Now().Hour()
	if hour >= 12 && hour < 18 {
		if !strings.Contains(output, "afternoon") {
			t.Errorf("expected 'afternoon' in greeting for hour %d, got: %s", hour, output)
		}
	} else if hour >= 18 {
		if !strings.Contains(output, "evening") {
			t.Errorf("expected 'evening' in greeting for hour %d, got: %s", hour, output)
		}
	} else {
		if !strings.Contains(output, "morning") {
			t.Errorf("expected 'morning' in greeting for hour %d, got: %s", hour, output)
		}
	}
}

// Test Cobra Command Definitions and Structure
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

// Test Flags Setup
func TestCLIFlags(t *testing.T) {
	flags := []string{"noweather", "noquotes", "noevents", "mini"}

	for _, flagName := range flags {
		flag := rootCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("expected CLI flag '--%s' to be defined", flagName)
		}
	}
}

// Test Application Execution Pipeline in Mini Mode
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

// Test Application Execution Pipeline in Full Mode
func TestRunApplication_FullMode(t *testing.T) {
	_ = setupTestConfig(t) // Seeds config file so wizard is bypassed

	opts := CLIOptions{
		Mini:      false,
		NoWeather: true, // Off to avoid network calls during main test
		NoQuotes:  true,
		NoEvents:  true,
	}

	output := captureStdout(func() {
		_ = runApplication(opts)
	})

	if !strings.Contains(output, "TestUser") {
		t.Errorf("expected output to contain user name 'TestUser', got: %s", output)
	}
}
