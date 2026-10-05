package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/inestrivino/bonjour/internal/config"
	"github.com/inestrivino/bonjour/internal/modules/events"
	"github.com/inestrivino/bonjour/internal/modules/quotes"
	"github.com/inestrivino/bonjour/internal/modules/weather"
	"github.com/inestrivino/bonjour/internal/ui"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// COBRA CONFIGURATION FOR FLAGS, ETC

type CLIOptions struct {
	// Module flags
	NoWeather bool
	NoQuotes  bool
	NoEvents  bool

	// Modes
	Mini bool
}

var (
	cliOpts CLIOptions
	version = "0.0.1"
)

// COBRA COMMANDS

var rootCmd = &cobra.Command{
	Use:     "bonjour",
	Short:   "bonjour is a terminal helper to start your day right",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runApplication(cliOpts)
	},
}

// Allow mocking in tests to prevent /dev/tty errors or terminal getting stuck by the form
var runConfigWizard = config.ConfigWizard

// configCmd represents the 'bonjour config' subcommand
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Open the configuration wizard to edit your settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadOrRunWizard()
		if err != nil {
			return fmt.Errorf("failed to load existing config: %w", err)
		}

		theme := ui.NewTheme(cfg.Dashboard.Theme)
		return runConfigWizard(theme)
	},
}

// Allow mocking in tests to prevent /dev/tty errors or terminal getting stuck by the form
var runEventsWizard = events.EventsWizard

// eventsCmd represents the 'bonjour events' subcommand
var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Open the events wizard to add or delete events",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadOrRunWizard()
		if err != nil {
			return fmt.Errorf("failed to load existing config: %w", err)
		}

		theme := ui.NewTheme(cfg.Dashboard.Theme)
		return runEventsWizard(theme)
	},
}

// Functions to run the app

// Main execution logic
func runApplication(opts CLIOptions) error {
	// Obtain the user's configuration
	cfg, err := config.LoadOrRunWizard()
	if err != nil {
		log.Fatalf("error running wizard: %v", err)
	}

	// Obtain the user's theme
	theme := ui.NewTheme(cfg.Dashboard.Theme)

	// Obtain the modules to be shown
	showQuotes, showWeather, showEvents := determineModulesToShow(opts, cfg)

	// Render the components according to the user's choices
	return renderComponents(showQuotes, showWeather, showEvents, opts.Mini, theme, cfg)
}

// Cobra initialization for commands and flags
func init() {
	// Subcommands
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(eventsCmd)

	// Module flags
	rootCmd.Flags().BoolVar(&cliOpts.NoWeather, "noweather", false, "don't show weather module")
	rootCmd.Flags().BoolVar(&cliOpts.NoQuotes, "noquotes", false, "don't show the quotes module")
	rootCmd.Flags().BoolVar(&cliOpts.NoEvents, "noevents", false, "don't show the events module")

	// Mini mode flag
	rootCmd.Flags().BoolVar(&cliOpts.Mini, "mini", false, "execute as a compact view")
}

// Main execution loop (Main function)
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// OTHER THINGS FOR THE MAIN SCRIPT

const asciiTitle = `
 __                                                    
/\ \                        __                         
\ \ \____    ___     ___   /\_\    ___   __  __  _ __  
 \ \ '` + "`" + `__` + "`" + `\  / __` + "`" + `\ /' _ ` + "`" + `\ \/\ \  / __` + "`" + `\/\ \/\ \/\` + "`" + `'__\
  \ \ \L\ \/\ \L\ \/\ \/\ \ \ \ \/\ \L\ \ \ \_\ \ \ \/ 
   \ \_,__/\ \____/\ \_\ \_\_\ \ \ \____/\ \____/\ \_\ 
    \/___/  \/___/  \/_/\/_/\ \_\ \/___/  \/___/  \/_/ 
                           \ \____/                    
                            \/___/`

// renderComponents is a helper function that takes in the relevant information from the config, theme, and user's module choices, to render the application's results
func renderComponents(showQuotes, showWeather, showEvents, mini bool, theme *ui.Theme, cfg *config.Config) error {
	// If the mini mode is activated, do not show the banner
	if !mini {
		fmt.Println(theme.Banner.Render(asciiTitle))
	}
	// Render the greeting
	greeting(cfg, theme, time.Now().Hour())

	// Fetch terminal width using x/term
	fd := int(os.Stdout.Fd())
	termWidth, _, err := term.GetSize(fd)
	if err != nil || termWidth <= 0 {
		termWidth = 90 // Fallback width for non-TTY environments
	}

	// Calculate target width (capped at 100 for proper layout ratios)
	totalWidth := termWidth - 4
	if totalWidth > 100 {
		totalWidth = 100
	}

	// We create a string object for the entire stdout result to show
	// The modules are appended to this
	var sections []string

	// The quote should take the entire width of the upper section
	if showQuotes {
		sections = append(sections, quotes.RenderQuote(theme, mini, totalWidth))
	}

	// The weather and events appear in side to side in the bottom section
	if showWeather && showEvents {
		if mini {
			// Mini mode: Render full-width and stack vertically
			wCard := weather.RenderWeatherData(cfg.User.Latitude, cfg.User.Longitude, cfg.User.City, theme, true, totalWidth)
			eCard := events.RenderEvents(theme, true, totalWidth)

			sections = append(sections, wCard, eCard)
		} else {
			const gap = 2
			availableWidth := totalWidth - gap
			leftWidth := availableWidth / 2
			rightWidth := availableWidth - leftWidth

			wCard := weather.RenderWeatherData(cfg.User.Latitude, cfg.User.Longitude, cfg.User.City, theme, false, leftWidth)
			eCard := events.RenderEvents(theme, false, rightWidth)

			gridRow := lipgloss.JoinHorizontal(lipgloss.Top, wCard, "  ", eCard)
			sections = append(sections, gridRow)
		}
	} else if showWeather {
		sections = append(sections, weather.RenderWeatherData(cfg.User.Latitude, cfg.User.Longitude, cfg.User.City, theme, mini, totalWidth))
	} else if showEvents {
		sections = append(sections, events.RenderEvents(theme, mini, totalWidth))
	}

	fmt.Println(lipgloss.JoinVertical(lipgloss.Left, sections...))
	return nil
}

// greeting takes in a Config type object and a theme type object, from which it renders a greeting based on the user's name, time of day, and theme
func greeting(cfg *config.Config, theme *ui.Theme, hour int) {
	greetingText := "morning"

	if hour >= 12 && hour < 18 {
		greetingText = "afternoon"
	} else if hour >= 18 {
		greetingText = "evening"
	}

	fmt.Println(theme.Subtitle.Render(fmt.Sprintf("Good %s, %s!\n", greetingText, cfg.User.Name)))
}

// Helper function to determine, given flags and the user's configuration, which modules to execute and which not
func determineModulesToShow(opts CLIOptions, cfg *config.Config) (bool, bool, bool) {
	var showQuotes bool
	var showWeather bool
	var showEvents bool

	// Whether to show quotes module or not
	if opts.NoQuotes {
		showQuotes = false
	} else {
		showQuotes = cfg.Dashboard.ShowQuotes
	}

	// Whether to show weather module or not
	if opts.NoWeather {
		showWeather = false
	} else if !(cfg.User.Latitude == 0 && cfg.User.Longitude == 0) {
		showWeather = cfg.Dashboard.ShowWeather
	}

	// Whether to show events module or not
	if opts.NoEvents {
		showEvents = false
	} else {
		showEvents = cfg.Dashboard.ShowEvents
	}

	return showQuotes, showWeather, showEvents
}
