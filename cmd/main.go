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

	"github.com/spf13/cobra"
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
	version = "1.0.0"
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
	Short: "open the configuration wizard to edit your settings",
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
	Short: "open the events wizard to add or delete events",
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
	// Load or create configuration
	cfg, err := config.LoadOrRunWizard()
	if err != nil {
		log.Fatalf("error running wizard: %v", err)
	}

	// Initialize the theme from the configuration
	theme := ui.NewTheme(cfg.Dashboard.Theme)

	// Render the banner
	if !opts.Mini {
		fmt.Println(theme.Banner.Render(asciiTitle))
	}

	// Render greeting
	greeting(cfg, theme, time.Now().Hour())

	// Render the active modules
	showQuotes, showWeather, showEvents := determineModulesToShow(opts, cfg)
	miniRender := opts.Mini
	if showQuotes {
		fmt.Println(quotes.RenderQuote(theme, miniRender))
	}
	if showWeather {
		fmt.Println(weather.RenderWeatherData(cfg.User.Latitude, cfg.User.Longitude, cfg.User.City, theme, miniRender))
	}
	if showEvents {
		fmt.Println(events.RenderEvents(theme, miniRender))
	}

	return nil
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
