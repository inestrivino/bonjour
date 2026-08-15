package events

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/inestrivino/bonjour/internal/config"
	"github.com/inestrivino/bonjour/internal/ui"
)

// Helper function to set up a cross-platform, isolated environment
func setupTestEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	t.Setenv("XDG_CONFIG_HOME", tmpDir) // Linux / BSD
	t.Setenv("HOME", tmpDir)            // macOS
	t.Setenv("APPDATA", tmpDir)         // Windows

	return tmpDir
}

// Helper to construct a mock UI Theme for testing LipGloss rendering
func mockTheme() *ui.Theme {
	return &ui.Theme{
		Title:     lipgloss.NewStyle(),
		Subtitle:  lipgloss.NewStyle(),
		Body:      lipgloss.NewStyle(),
		Muted:     lipgloss.Color("240"),
		ErrorText: lipgloss.NewStyle(),
		Card:      lipgloss.NewStyle(),
	}
}

// Test daysLeft logic across past, present, future, and invalid events
func TestDaysLeft(t *testing.T) {
	now := time.Now()
	todayStr := now.Format("2006-01-02")
	tomorrowStr := now.AddDate(0, 0, 1).Format("2006-01-02")
	yesterdayStr := now.AddDate(0, 0, -1).Format("2006-01-02")
	futureStr := now.AddDate(0, 0, 10).Format("2006-01-02")

	tests := []struct {
		name      string
		event     *config.EventConfig
		wantDays  int
		expectErr bool
	}{
		{
			name:      "Nil event pointer",
			event:     nil,
			wantDays:  0,
			expectErr: true,
		},
		{
			name:      "Invalid date string format",
			event:     &config.EventConfig{Title: "Bad Date", Date: "12/25/2026"},
			wantDays:  0,
			expectErr: true,
		},
		{
			name:      "Event happens today",
			event:     &config.EventConfig{Title: "Today Event", Date: todayStr},
			wantDays:  0,
			expectErr: false,
		},
		{
			name:      "Event happens tomorrow",
			event:     &config.EventConfig{Title: "Tomorrow Event", Date: tomorrowStr},
			wantDays:  1,
			expectErr: false,
		},
		{
			name:      "Event happened yesterday",
			event:     &config.EventConfig{Title: "Past Event", Date: yesterdayStr},
			wantDays:  -1,
			expectErr: false,
		},
		{
			name:      "Event 10 days in the future",
			event:     &config.EventConfig{Title: "Future Event", Date: futureStr},
			wantDays:  10,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			days, err := daysLeft(tt.event)

			if (err != nil) != tt.expectErr {
				t.Fatalf("daysLeft() error status = %v, expectErr = %v", err, tt.expectErr)
			}

			if !tt.expectErr && days != tt.wantDays {
				t.Errorf("daysLeft() = %d, want %d", days, tt.wantDays)
			}
		})
	}
}

// Test saveEvent helper persistence and slice appending
func TestSaveEvent(t *testing.T) {
	_ = setupTestEnv(t)
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("failed to resolve config path: %v", err)
	}

	// Initialize config with 1 base event
	initialConfig := &config.Config{
		Events: []config.EventConfig{
			{Title: "Existing Event", Date: "2026-10-10", WarningStart: "2026-10-01"},
		},
	}
	if err := config.SaveConfig(configPath, initialConfig); err != nil {
		t.Fatalf("failed to seed initial config: %v", err)
	}

	// Save a new event
	newEvent := &config.EventConfig{
		Title:        "New Birthday",
		Date:         "2026-12-25",
		WarningStart: "2026-12-18",
	}

	if err := saveEvent(newEvent); err != nil {
		t.Fatalf("saveEvent() returned unexpected error: %v", err)
	}

	// Read config back to verify appending works as expected
	updatedConfig, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to read updated config: %v", err)
	}

	if len(updatedConfig.Events) != 2 {
		t.Fatalf("expected 2 events in config, got %d", len(updatedConfig.Events))
	}

	if updatedConfig.Events[1].Title != "New Birthday" {
		t.Errorf("expected appended event title 'New Birthday', got '%s'", updatedConfig.Events[1].Title)
	}
}

// Test RenderEvents rendering variations and output structure
func TestRenderEvents(t *testing.T) {
	_ = setupTestEnv(t)
	theme := mockTheme()

	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("failed to resolve config path: %v", err)
	}

	now := time.Now()
	todayStr := now.Format("2006-01-02")
	future1 := now.AddDate(0, 0, 2).Format("2006-01-02")
	future2 := now.AddDate(0, 0, 5).Format("2006-01-02")
	farFutureWarning := now.AddDate(0, 0, 30).Format("2006-01-02")

	// Pre-populate configuration with various active/inactive events
	cfg := &config.Config{
		Events: []config.EventConfig{
			{Title: "Today Celebration", Date: todayStr, WarningStart: todayStr},
			{Title: "Near Event", Date: future1, WarningStart: todayStr},
			{Title: "Later Event", Date: future2, WarningStart: todayStr},
			{Title: "Hidden Event", Date: "2026-12-31", WarningStart: farFutureWarning}, // Warning start in the future
			{Title: "Expired Event", Date: "2020-01-01", WarningStart: "2019-12-01"},    // Past event
		},
	}

	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("failed to save config seed: %v", err)
	}

	t.Run("Standard Rendering", func(t *testing.T) {
		output := RenderEvents(theme, false)

		if !strings.Contains(output, "Upcoming Events") {
			t.Error("expected output to contain 'Upcoming Events' header")
		}
		if !strings.Contains(output, "Today Celebration") {
			t.Error("expected output to display active today event")
		}
		if !strings.Contains(output, "Near Event") {
			t.Error("expected output to display near future event")
		}
		if strings.Contains(output, "Hidden Event") {
			t.Error("expected output to omit event with future WarningStart date")
		}
		if strings.Contains(output, "Expired Event") {
			t.Error("expected output to omit past events")
		}
	})

	t.Run("Mini Rendering", func(t *testing.T) {
		output := RenderEvents(theme, true)

		// Mini rendering highlights the single nearest active event
		if !strings.Contains(output, "Today Celebration") {
			t.Errorf("expected mini output to show 'Today Celebration', got: %s", output)
		}
		if !strings.Contains(output, "Today!") {
			t.Errorf("expected mini output to show 'Today!', got: %s", output)
		}
	})
}

// Test RenderEvents empty state behavior
func TestRenderEvents_EmptyState(t *testing.T) {
	_ = setupTestEnv(t)
	theme := mockTheme()

	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("failed to resolve config path: %v", err)
	}

	// Create config with empty events
	cfg := &config.Config{Events: []config.EventConfig{}}
	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("failed to seed empty config: %v", err)
	}

	t.Run("Standard View Empty", func(t *testing.T) {
		out := RenderEvents(theme, false)
		if !strings.Contains(out, "No upcoming events scheduled.") {
			t.Errorf("expected empty state standard message, got: %s", out)
		}
	})

	t.Run("Mini View Empty", func(t *testing.T) {
		out := RenderEvents(theme, true)
		if !strings.Contains(out, "No events scheduled") {
			t.Errorf("expected empty state mini message, got: %s", out)
		}
	})
}

// Test RenderEvents error fallback state when config cannot be loaded
func TestRenderEvents_MissingConfigError(t *testing.T) {
	tmpDir := setupTestEnv(t)
	theme := mockTheme()

	// Direct config path to an uncreatable/invalid directory path to force error
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "non_existent_path"))

	out := RenderEvents(theme, false)
	if !strings.Contains(out, "Events Unavailable") {
		t.Errorf("expected error card output 'Events Unavailable', got: %s", out)
	}
}

// Test limit handling when more than 5 events are active
func TestRenderEvents_TopFiveLimit(t *testing.T) {
	_ = setupTestEnv(t)
	theme := mockTheme()

	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("failed to resolve path: %v", err)
	}

	now := time.Now()
	var events []config.EventConfig

	// Generate 7 active events
	for i := 1; i <= 7; i++ {
		events = append(events, config.EventConfig{
			Title:        "Event " + string(rune('0'+i)),
			Date:         now.AddDate(0, 0, i).Format("2006-01-02"),
			WarningStart: now.Format("2006-01-02"),
		})
	}

	cfg := &config.Config{Events: events}
	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	out := RenderEvents(theme, false)

	// Ensure top 5 are displayed and 6th and 7th are truncated
	if !strings.Contains(out, "Event 1") || !strings.Contains(out, "Event 5") {
		t.Error("expected first 5 events to be rendered")
	}

	if strings.Contains(out, "Event 6") || strings.Contains(out, "Event 7") {
		t.Error("expected events beyond top 5 to be truncated")
	}
}
