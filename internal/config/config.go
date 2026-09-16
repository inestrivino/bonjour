package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/inestrivino/bonjour/internal/modules/weather"
)

type UserConfig struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"city"`
}

type DashboardConfig struct {
	Theme       string `json:"theme"`
	ShowWeather bool   `json:"show_weather"`
	ShowQuotes  bool   `json:"show_quotes"`
	ShowEvents  bool   `json:"show_events"`
}

type EventConfig struct {
	Date         string `json:"date"`
	Title        string `json:"title"`
	WarningStart string `json:"warningstart"`
}

type Config struct {
	User      UserConfig      `json:"user"`
	Dashboard DashboardConfig `json:"dashboard"`
	Events    []EventConfig   `json:"events"`
}

// GetConfigPath returns the path to the user configuration
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "bonjour", "config.json"), nil
}

// SaveConfig saves a Config type object as a JSON in the user configuration file
func SaveConfig(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, bytes, 0644)
}

// LoadConfig loads the information within the user configuration file into a Config type object
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

// Helper function to validate that the user's name is not empty
func ValidateName(str string) error {
	if len(strings.TrimSpace(str)) == 0 {
		return errors.New("name can't be empty")
	}
	return nil
}

// Allow mocking in tests to prevent /dev/tty errors or terminal getting stuck by the form
var resolveLocation = weather.ResolveLocation

// Given a Config object, a city name, a country, and administration, sets the correct Latitude, Longitude and City in the Config
func ResolveAndApplyLocation(cfg *Config, city, country, admin string) error {
	cleanCity := strings.TrimSpace(city)
	if cleanCity == "" {
		cfg.User.Latitude = 0
		cfg.User.Longitude = 0
		cfg.User.City = ""
		return nil
	}

	loc, err := resolveLocation(cleanCity, country, admin)
	if err != nil {
		return err
	}

	cfg.User.Latitude = loc.Latitude
	cfg.User.Longitude = loc.Longitude
	cfg.User.City = cleanCity
	return nil
}

// ApplySelectedModules sets the modules in the main Config object to false or true depending on what was sent
func ApplySelectedModules(cfg *Config, selectedModules []string) {
	cfg.Dashboard.ShowWeather = false
	cfg.Dashboard.ShowQuotes = false
	cfg.Dashboard.ShowEvents = false

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
}

// LoadOrRunWizard loads the user configuration from its file and then determines wheter the initial wizard needs to be run or execution can continue as usual
func LoadOrRunWizard() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("error while trying to obtain the configuration directory: %w", err)
	}

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		cfg, err := initialWizard()
		if err != nil {
			return nil, err
		}

		if err := SaveConfig(configPath, cfg); err != nil {
			return nil, fmt.Errorf("error saving user configuration: %w", err)
		}

		fmt.Println("\nsuccessfully saved preferences in: ", configPath)

		return cfg, nil
	}

	return LoadConfig(configPath)
}
