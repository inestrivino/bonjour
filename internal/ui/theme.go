package ui

// Package ui implements the functionalities to manage the app's consistent styling and giving the user styling choices

import (
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Theme is the collection of styling choices that determine how the app's UI displays
type Theme struct {
	Primary   lipgloss.TerminalColor //primary color
	Secondary lipgloss.TerminalColor //secondary color
	Muted     lipgloss.TerminalColor //a muted color
	Error     lipgloss.TerminalColor //the error color

	Banner    lipgloss.Style //banner style
	Card      lipgloss.Style //card style
	Title     lipgloss.Style //title style
	Subtitle  lipgloss.Style //subtitle style
	Body      lipgloss.Style //body style
	ErrorText lipgloss.Style //style for error texts

	HuhTheme *huh.Theme
}

// NewTheme takes the name of a theme and applies it in the CLI UI. It returns the Theme object.
func NewTheme(themeName string) *Theme {
	// If the user has applied the "NO_COLOR" global environment variable, the UI will be colorless
	// as per https://no-color.org/
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
		lipgloss.SetColorProfile(termenv.Ascii)
		return &Theme{
			HuhTheme:  huh.ThemeBase(),
			Banner:    lipgloss.NewStyle().Bold(true).MarginBottom(1),
			Card:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2),
			Title:     lipgloss.NewStyle().Bold(true),
			Subtitle:  lipgloss.NewStyle().Italic(true),
			Body:      lipgloss.NewStyle(),
			ErrorText: lipgloss.NewStyle().Bold(true),
		}
	}

	// We select a pre-built theme from the huh library
	var huhTheme *huh.Theme
	switch themeName {
	case "dracula":
		huhTheme = huh.ThemeDracula()
	case "catppuccin":
		huhTheme = huh.ThemeCatppuccin()
	case "base16":
		huhTheme = huh.ThemeBase16()
	default:
		huhTheme = huh.ThemeCharm()
	}

	// We extract the palette and styling from the theme
	t := &Theme{
		Primary:   huhTheme.Focused.Base.GetBorderLeftForeground(),
		Secondary: huhTheme.Focused.Title.GetForeground(),
		Muted:     huhTheme.Blurred.Title.GetForeground(),
		Error:     huhTheme.Focused.ErrorMessage.GetForeground(),
		HuhTheme:  huhTheme,
	}
	t.Banner = lipgloss.NewStyle().Foreground(t.Secondary).Bold(true).MarginBottom(1)
	t.Card = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.Primary).Padding(1, 2)
	t.Title = lipgloss.NewStyle().Foreground(t.Secondary).Bold(true)
	t.Subtitle = lipgloss.NewStyle().Foreground(t.Muted).Italic(true)
	t.Body = lipgloss.NewStyle()
	t.ErrorText = lipgloss.NewStyle().Foreground(t.Error).Bold(true)

	return t
}
