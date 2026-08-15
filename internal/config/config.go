package config

// Package config implements the functionalities to help the user set the preferences in the application's configuration.
//
// This includes a wizard that helps users with the initial setup upon installation, and change it.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/inestrivino/bonjour/internal/modules/weather"
	"github.com/inestrivino/bonjour/internal/ui"
)

// A UserConfig is the user's preferences regarding themselves.
// Preferred name, preferred location.
type UserConfig struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"city"`
}

// A DashboardConfig is the user's preferences regarding the dashboard.
// Show certain modules (such as weather, quotes...) or not.
type DashboardConfig struct {
	Theme       string `json:"theme"`
	ShowWeather bool   `json:"show_weather"`
	ShowQuotes  bool   `json:"show_quotes"`
	ShowEvents  bool   `json:"show_events"`
}

// An EventConfig is an event that the user wants to be reminded of.
// The date the event happens, the title of the event, and from which date forward they want to be warned of it.
type EventConfig struct {
	Date         string `json:"date"`
	Title        string `json:"title"`
	WarningStart string `json:"warningstart"`
}

// A Config is the entire user configuration of the app.
// It includes the UserConfig, DashboardConfig, EventConfig from which to receive information.
type Config struct {
	User      UserConfig      `json:"user"`
	Dashboard DashboardConfig `json:"dashboard"`
	Events    []EventConfig   `json:"events"`
}

// GetConfigPath returns a string with the path where the Config is stored. It may also return an error.
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appCOnfigDir := filepath.Join(configDir, "bonjour", "config.json")
	return appCOnfigDir, nil
}

// saveConfig saves the Config into a readable JSON file.
// It returns an error.
func SaveConfig(path string, cfg *Config) error {
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

// loadConfig reads the JSON config file.
// It returns a Config object and an error.
func LoadConfig(path string) (*Config, error) {
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

// LoadOrRunWizard checks whether there is a configuration file or not. If there is not, it launches the configuration wizard for first-time users.
// It returns a Config type object.
func LoadOrRunWizard() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("Error while trying to obtain the configuration directory: %w", err)
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
		if err := SaveConfig(configPath, cfg); err != nil {
			return nil, fmt.Errorf("Error saving user configuration: %w", err)
		}

		fmt.Println("\nSuccessfully saved preferences in: ", configPath)

		return cfg, nil
	}

	return LoadConfig(configPath)
}

// initialWizard presents a form of questions for the user to answer, which will be used in the Config.
// It returns a Config object and an error.
func initialWizard() (*Config, error) {
	// We create a new Config object (values initialized to 0) and return the pointer to it
	// We use a pointer here to avoid dealing with passing copies back and forth
	cfg := &Config{}
	selectedModules := []string{"weather", "quotes", "events"}

	// We initialize the charm theme as default for the initial wizard
	initialTheme := ui.NewTheme("charm")

	// variables to perform the geolocation api call for weather services
	var (
		inputCity    string
		inputCountry string
		inputAdmin   string
	)

	form := huh.NewForm(
		huh.NewGroup(
			// User name
			huh.NewInput().
				Title("What would you like to be called?").
				Placeholder("John Doe").
				Value(&cfg.User.Name).
				Validate(func(str string) error {
					if len(str) == 0 {
						return errors.New("Name can't be empty")
					}
					return nil
				}),
		),
		// Location conf
		huh.NewGroup(
			huh.NewInput().
				Title("City Name").
				Description("Optional. Leave blank to disable weather features.").
				Placeholder("Paris").
				Value(&inputCity),

			huh.NewInput().
				Title("Country").
				Description("Optional (e.g., Spain, France, UK, ES)").
				Placeholder("France").
				Value(&inputCountry),

			huh.NewInput().
				Title("Administrative Area").
				Description("Optional (e.g., Texas, Île-de-France, Catalunya)").
				Placeholder("Île-de-France").
				Value(&inputAdmin).
				Validate(func(_ string) error {
					// Validate the combination when the user hits Enter on the final input
					if strings.TrimSpace(inputCity) == "" {
						return nil // Skip weather setup if city is left blank
					}

					loc, err := weather.ResolveLocation(inputCity, inputCountry, inputAdmin)
					if err != nil {
						return err // Returns error allowing user to adjust inputs
					}

					// Save matched results directly into config
					cfg.User.Latitude = loc.Latitude
					cfg.User.Longitude = loc.Longitude
					cfg.User.City = inputCity

					return nil
				}),
		),
		huh.NewGroup(
			// DashboardConfig
			huh.NewMultiSelect[string]().
				Title("Select the dashboard modules you want to activate:").
				Description("Use the [Space] key to toggle or untoggle modules").
				Options(
					huh.NewOption("Weather", "weather"),
					huh.NewOption("Quotes", "quotes"),
					huh.NewOption("Events", "events"),
				).
				Value(&selectedModules),

			huh.NewSelect[string]().
				Title("Select UI Theme:").
				Options(
					huh.NewOption("Charm (Default)", "charm"),
					huh.NewOption("Dracula", "dracula"),
					huh.NewOption("Catppuccin", "catppuccin"),
					huh.NewOption("Base16", "base16"),
				).
				Value(&cfg.Dashboard.Theme),
		),
	).WithTheme(initialTheme.HuhTheme)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	// Map selections to bool values.
	// For iteration over a range of selectedModules, for each one mod receives the string item ("weather") and sets the corresponding dashboard element to true.
	for _, mod := range selectedModules {
		switch mod {
		case "weather":
			cfg.Dashboard.ShowWeather = true
		case "quotes":
			cfg.Dashboard.ShowQuotes = true
		case "events":
			cfg.Dashboard.ShowEvents = true
		}
	}

	// We initialize the list of events as an empty list.
	cfg.Events = make([]EventConfig, 0)

	return cfg, nil
}

// ConfigWizard is a menu function to help the user alter their configuration.
func ConfigWizard(currentTheme *ui.Theme) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("could not resolve config path: %w", err)
	}

	// Load existing configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("Error while loading the previous configuration: %w", err)
	}

	// User selects which part of the config to edit
	var confToEdit string
	firstForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select one to edit:").
				Options(
					huh.NewOption("Change username", "username"),
					huh.NewOption("Change or delete location", "location"),
					huh.NewOption("Dashboard configuration", "dashboard"),
					huh.NewOption("Go back", "back"),
				).
				Value(&confToEdit),
		),
	).WithTheme(currentTheme.HuhTheme)

	if err := firstForm.Run(); err != nil {
		return err
	}

	// Based on the user selection one form or the other is shown
	switch confToEdit {
	case "username":
		userNameForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("What would you like to be called?").
					Placeholder(config.User.Name).
					Value(&config.User.Name).
					Validate(func(str string) error {
						if len(str) == 0 {
							return errors.New("Name can't be empty")
						}
						return nil
					}),
			),
		).WithTheme(currentTheme.HuhTheme)

		if err := userNameForm.Run(); err != nil {
			return err
		}

	case "location":
		var inputCity, inputCountry, inputAdmin string
		locationForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("City Name").
					Description("Optional. Leave blank to disable weather features.").
					Placeholder("Paris").
					Value(&inputCity),

				huh.NewInput().
					Title("Country").
					Description("Optional (e.g., Spain, France, UK, ES)").
					Placeholder("France").
					Value(&inputCountry),

				huh.NewInput().
					Title("Administrative Area").
					Description("Optional (e.g., Texas, Île-de-France, Catalunya)").
					Placeholder("Île-de-France").
					Value(&inputAdmin).
					Validate(func(_ string) error {
						cleanCity := strings.TrimSpace(inputCity)

						// If city is empty, clear out the location data in config
						if cleanCity == "" {
							config.User.Latitude = 0
							config.User.Longitude = 0
							config.User.City = ""
							return nil
						}

						loc, err := weather.ResolveLocation(cleanCity, inputCountry, inputAdmin)
						if err != nil {
							return err // Returns error allowing user to adjust inputs
						}

						// Save matched results directly into config
						config.User.Latitude = loc.Latitude
						config.User.Longitude = loc.Longitude
						config.User.City = cleanCity

						return nil
					}),
			),
		).WithTheme(currentTheme.HuhTheme)

		if err := locationForm.Run(); err != nil {
			return err
		}
	case "dashboard":
		var selectedModules []string
		if config.Dashboard.ShowWeather {
			selectedModules = append(selectedModules, "weather")
		}
		if config.Dashboard.ShowQuotes {
			selectedModules = append(selectedModules, "quotes")
		}
		if config.Dashboard.ShowEvents {
			selectedModules = append(selectedModules, "events")
		}

		dashboardForm := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select the dashboard modules you want to activate:").
					Description("Use the [Space] key to toggle or untoggle modules").
					Options(
						huh.NewOption("Weather", "weather"),
						huh.NewOption("Quotes", "quotes"),
						huh.NewOption("Events", "events"),
					).
					Value(&selectedModules),

				huh.NewSelect[string]().
					Title("Select UI Theme:").
					Options(
						huh.NewOption("Charm (Default)", "charm"),
						huh.NewOption("Dracula", "dracula"),
						huh.NewOption("Catppuccin", "catppuccin"),
						huh.NewOption("Base16", "base16"),
					).
					Value(&config.Dashboard.Theme),
			),
		).WithTheme(currentTheme.HuhTheme)

		if err := dashboardForm.Run(); err != nil {
			return err
		}

		// Reset booleans before applying the updated selection state
		config.Dashboard.ShowWeather = false
		config.Dashboard.ShowQuotes = false
		config.Dashboard.ShowEvents = false

		for _, mod := range selectedModules {
			switch mod {
			case "weather":
				config.Dashboard.ShowWeather = true
			case "quotes":
				config.Dashboard.ShowQuotes = true
			case "events":
				config.Dashboard.ShowEvents = true
			}
		}

	case "back":
		return nil
	}

	// Persist changes back to disk
	if err := SaveConfig(configPath, config); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	fmt.Println("successfully updated the configuration!")

	return nil
}
