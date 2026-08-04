package config

// Package config implements the functionalities to help the user set the preferences in the application's configuration.
//
// This includes a wizard that helps users with the initial setup upon installation, methods to save and retrieve the configuration, etc.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
)

// A UserConfig is the user's preferences regarding themselves
// Preferred name, preferred location and preferred language for the app to use
type UserConfig struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Language string `json:"language"`
}

// A DashboardConfig is the user's preferences regarding the dashboard
// Show certain modules (such as weather, quotes, rss...) or not
// TODO: ALLOW CHOOSE THEME?
type DashboardConfig struct {
	ShowWeather bool `json:"show_weather"`
	ShowQuotes  bool `json:"show_quotes"`
	ShowRSS     bool `json:"show_rss"`
}

// An EventConfig is an event that the user wants to be reminded of
// The date the event happens, the title of the event, and from which date forward they want to be warned of it
type EventConfig struct {
	Date         string `json:"date"`
	Title        string `json:"title"`
	WarningStart string `json:"warningstart"`
}

// A Config is the entire user configuration of the app
// It includes the UserConfig, DashboardConfig, EventConfig, and a list of RSSFeeds from which to receive information
type Config struct {
	User      UserConfig      `json:"user"`
	Dashboard DashboardConfig `json:"dashboard"`
	RSSFeeds  []string        `json:"rss_feeds"`
	Events    []EventConfig   `json:"events"`
}

// GetConfigPath returns a string with the path where the Config is stored. It may also return an error.
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "bonjour", "config.json"), nil
}

// LoadOrRunWizard checks whether there is a configuration file or not. If there is not, it launches the configuration wizard for first-time users.
// It returns a Config type object
func LoadOrRunWizard() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("Error while trying to obtain the home directory: %w", err)
	}

	//we check whether the config file exists
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		//if there is no config file then we run the first-time installation wizard
		//cfg is type Config
		cfg, err := initialWizard()
		if err != nil {
			return nil, err
		}

		// We call the function to save the configuration. It may only return an error. If it does return an error then we make it known. Otherwise we assume success.
		if err := saveConfig(configPath, cfg); err != nil {
			return nil, fmt.Errorf("Error saving user configuration: %w", err)
		}

		fmt.Println("\nSuccessfully saved preferences in: ", configPath)

		return cfg, nil
	}

	return loadConfig(configPath)
}

// initialWizard presents a form of questions for the user to answer, which will be used in the Config.
// It returns a Config object and an error
func initialWizard() (*Config, error) {
	// We create a new Config object (values initialized to 0) and return the pointer to it
	// We use a pointer here to avoid dealing with passing copies back and forth
	cfg := &Config{}
	selectedModules := []string{"weather", "quotes", "rss"}

	form := huh.NewForm(
		huh.NewGroup(
			// UserConfig
			huh.NewInput().
				Title("What would you like to be called?").
				Placeholder("María").
				Value(&cfg.User.Name).
				Validate(func(str string) error {
					if len(str) == 0 {
						return errors.New("Name can't be emtpy")
					}
					return nil
				}),

			huh.NewInput().
				Title("What is your location?").
				Description("Optional. Leaving this blank disables weather features").
				Placeholder("Madrid, Spain").
				Value(&cfg.User.Location),

			huh.NewSelect[string]().
				Title("Select your preferred language:").
				Options(
					huh.NewOption("English", "en"),
					huh.NewOption("Español", "es"),
					huh.NewOption("Français", "fr"),
				).
				Value(&cfg.User.Language),

			// DashboardConfig
			huh.NewMultiSelect[string]().
				Title("Select the dashboard modules you want to activate:").
				Description("Use the [Space] key to toggle or untoggle modules").
				Options(
					huh.NewOption("Weather", "weather"),
					huh.NewOption("Quotes", "quotes"),
					huh.NewOption("News and RSS", "rss"),
				).
				Value(&selectedModules),
		),
	).WithTheme(huh.ThemeCharm())

	err := form.Run()
	if err != nil {
		return nil, err
	}

	// Map selections to bool values
	// For iteration over a range of selectedModules, for each one mod receives the string item ("weather") and sets the corresponding dashboard element to true
	for _, mod := range selectedModules {
		switch mod {
		case "weather":
			cfg.Dashboard.ShowWeather = true
		case "quotes":
			cfg.Dashboard.ShowQuotes = true
		case "rss":
			cfg.Dashboard.ShowRSS = true
		}
	}

	return cfg, nil
}

// saveConfig saves the Config into a readable JSON file
// It returns an error
func saveConfig(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Whitespace application to ensure JSON legibility
	bytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, bytes, 0644)
}

// loadConfig reads the JSON config file
// It returns a Config object and an error
func loadConfig(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// TODO: CREATE SOME SORT OF WIZARD THAT ALLOWS THE USER TO ALTER THIS CONFIGS LATER ON
