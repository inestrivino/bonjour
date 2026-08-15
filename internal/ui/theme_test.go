package ui

import (
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// THEME CREATION & INITIALIZATION TESTS

func TestNewTheme_SupportedThemes(t *testing.T) {
	// Ensure NO_COLOR isn't interfering with standard color profile tests
	os.Unsetenv("NO_COLOR")

	tests := []struct {
		name      string
		themeName string
	}{
		{"Dracula Theme", "dracula"},
		{"Catppuccin Theme", "catppuccin"},
		{"Base16 Theme", "base16"},
		{"Default Charm Theme (Explicit)", "charm"},
		{"Default Fallback Theme (Unknown input)", "unknown-theme-xyz"},
		{"Empty Theme Name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := NewTheme(tt.themeName)

			if theme == nil {
				t.Fatalf("NewTheme(%q) returned nil", tt.themeName)
			}
			if theme.HuhTheme == nil {
				t.Errorf("NewTheme(%q).HuhTheme is nil", tt.themeName)
			}

			// Validate essential colors extracted from the selected palette
			if theme.Primary == nil {
				t.Errorf("NewTheme(%q).Primary color is nil", tt.themeName)
			}
			if theme.Secondary == nil {
				t.Errorf("NewTheme(%q).Secondary color is nil", tt.themeName)
			}
			if theme.Muted == nil {
				t.Errorf("NewTheme(%q).Muted color is nil", tt.themeName)
			}
			if theme.Error == nil {
				t.Errorf("NewTheme(%q).Error color is nil", tt.themeName)
			}

			// Validate LipGloss style initializations
			assertStyleConfigured(t, "Banner", theme.Banner)
			assertStyleConfigured(t, "Card", theme.Card)
			assertStyleConfigured(t, "Title", theme.Title)
			assertStyleConfigured(t, "Subtitle", theme.Subtitle)
			assertStyleConfigured(t, "ErrorText", theme.ErrorText)
		})
	}
}

// EDGE CASES & ENVIRONMENT CONTROLS

func TestNewTheme_NoColorEnvironmentVariable(t *testing.T) {
	// Set NO_COLOR compliance flag
	t.Setenv("NO_COLOR", "1")

	// Reset color profile state after test completes
	defer lipgloss.SetColorProfile(termenv.ColorProfile())

	theme := NewTheme("dracula")

	if theme == nil {
		t.Fatal("NewTheme returned nil when NO_COLOR was set")
	}

	// NO_COLOR branch should leave base palette fields unassigned/nil
	if theme.Primary != nil || theme.Secondary != nil || theme.Muted != nil || theme.Error != nil {
		t.Errorf("Expected nil palette colors under NO_COLOR, got Primary: %v, Secondary: %v", theme.Primary, theme.Secondary)
	}

	if theme.HuhTheme == nil {
		t.Error("Expected HuhTheme to be set to BaseTheme under NO_COLOR")
	}

	// Verify standard styles render without crash or color bindings
	renderedCard := theme.Card.Render("test card content")
	if renderedCard == "" {
		t.Error("Card style rendered empty string")
	}
}

func TestTheme_RenderOutputSanity(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	theme := NewTheme("catppuccin")
	sampleText := "Hello World"

	t.Run("Title Rendering", func(t *testing.T) {
		out := theme.Title.Render(sampleText)
		if out == "" {
			t.Error("Title rendered empty string")
		}
	})

	t.Run("Card Component Frame", func(t *testing.T) {
		out := theme.Card.Render(sampleText)
		if out == "" {
			t.Error("Card component rendered empty string")
		}
	})

	t.Run("Error Text Formatting", func(t *testing.T) {
		out := theme.ErrorText.Render("Critical Failure")
		if out == "" {
			t.Error("ErrorText style rendered empty string")
		}
	})
}

// HELPERS

func assertStyleConfigured(t *testing.T, styleName string, style lipgloss.Style) {
	t.Helper()
	// An initialized LipGloss style rendered with empty string shouldn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Style %s failed rendering check: %v", styleName, r)
		}
	}()
	_ = style.Render("")
}
