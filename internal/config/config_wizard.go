package config

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/inestrivino/bonjour/internal/ui"
)

func initialWizard() (*Config, error) {
	cfg := &Config{}
	selectedModules := []string{"weather", "quotes", "events"}
	initialTheme := ui.NewTheme("charm")

	var (
		inputCity    string
		inputCountry string
		inputAdmin   string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("what would you like to be called?").
				Placeholder("John Doe").
				Value(&cfg.User.Name).
				Validate(ValidateName),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("city name").
				Description("optional: leave blank to disable weather features.").
				Placeholder("Paris").
				Value(&inputCity),

			huh.NewInput().
				Title("country").
				Description("optional (e.g., Spain, France, UK, ES)").
				Placeholder("France").
				Value(&inputCountry),

			huh.NewInput().
				Title("administrative area").
				Description("optional (e.g., Texas, Île-de-France, Catalunya)").
				Placeholder("Île-de-France").
				Value(&inputAdmin).
				Validate(func(_ string) error {
					return ResolveAndApplyLocation(cfg, inputCity, inputCountry, inputAdmin)
				}),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("select the dashboard modules you want to activate:").
				Description("use the [Space] key to toggle or untoggle modules").
				Options(
					huh.NewOption("weather", "weather"),
					huh.NewOption("quotes", "quotes"),
					huh.NewOption("events", "events"),
				).
				Value(&selectedModules),

			huh.NewSelect[string]().
				Title("select UI Theme:").
				Options(
					huh.NewOption("charm (default)", "charm"),
					huh.NewOption("dracula", "dracula"),
					huh.NewOption("catppuccin", "catppuccin"),
					huh.NewOption("base16", "base16"),
				).
				Value(&cfg.Dashboard.Theme),
		),
	).WithTheme(initialTheme.HuhTheme)

	if err := form.Run(); err != nil {
		return nil, err
	}

	ApplySelectedModules(cfg, selectedModules)
	cfg.Events = make([]EventConfig, 0)

	return cfg, nil
}

func ConfigWizard(currentTheme *ui.Theme) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("could not resolve config path: %w", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error while loading the previous configuration: %w", err)
	}

	var confToEdit string
	firstForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("select one to edit:").
				Options(
					huh.NewOption("change username", "username"),
					huh.NewOption("change or delete location", "location"),
					huh.NewOption("dashboard configuration", "dashboard"),
					huh.NewOption("go back", "back"),
				).
				Value(&confToEdit),
		),
	).WithTheme(currentTheme.HuhTheme)

	if err := firstForm.Run(); err != nil {
		return err
	}

	switch confToEdit {
	case "username":
		userNameForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("what would you like to be called?").
					Placeholder(config.User.Name).
					Value(&config.User.Name).
					Validate(ValidateName),
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
					Title("city Name").
					Description("optional: leave blank to disable weather features.").
					Placeholder("Paris").
					Value(&inputCity),

				huh.NewInput().
					Title("country").
					Description("optional (e.g., Spain, France, UK, ES)").
					Placeholder("France").
					Value(&inputCountry),

				huh.NewInput().
					Title("administrative area").
					Description("optional (e.g., Texas, Île-de-France, Catalunya)").
					Placeholder("Île-de-France").
					Value(&inputAdmin).
					Validate(func(_ string) error {
						return ResolveAndApplyLocation(config, inputCity, inputCountry, inputAdmin)
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
					Title("select the dashboard modules you want to activate:").
					Description("use the [Space] key to toggle or untoggle modules").
					Options(
						huh.NewOption("weather", "weather"),
						huh.NewOption("quotes", "quotes"),
						huh.NewOption("events", "events"),
					).
					Value(&selectedModules),

				huh.NewSelect[string]().
					Title("Select UI Theme:").
					Options(
						huh.NewOption("charm (Default)", "charm"),
						huh.NewOption("dracula", "dracula"),
						huh.NewOption("catppuccin", "catppuccin"),
						huh.NewOption("base16", "base16"),
					).
					Value(&config.Dashboard.Theme),
			),
		).WithTheme(currentTheme.HuhTheme)

		if err := dashboardForm.Run(); err != nil {
			return err
		}

		ApplySelectedModules(config, selectedModules)

	case "back":
		return nil
	}

	if err := SaveConfig(configPath, config); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	fmt.Println("successfully updated the configuration!")

	return nil
}
