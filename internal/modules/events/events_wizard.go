package events

import (
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/inestrivino/bonjour/internal/config"
	"github.com/inestrivino/bonjour/internal/ui"
)

func newEvent(currentTheme *ui.Theme) error {
	var (
		title        string
		dateStr      string
		warningStart string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("event title").
				Placeholder("mom's birthday").
				Value(&title).
				Validate(ValidateEventTitle),

			huh.NewInput().
				Title("event date").
				Description("format: YYYY-MM-DD").
				Placeholder("2026-12-25").
				Value(&dateStr).
				Validate(ValidateEventDate),

			huh.NewInput().
				Title("warning start date").
				Description("from which date forward do you want to be reminded? (YYYY-MM-DD). leave empty to start today.").
				Placeholder("2026-12-18").
				Value(&warningStart).
				Validate(func(str string) error {
					return ValidateWarningDate(str, dateStr, time.Now())
				}),
		),
	).WithTheme(currentTheme.HuhTheme)

	if err := form.Run(); err != nil {
		return err
	}

	event := CreateEventObject(title, dateStr, warningStart)

	if err := saveEvent(event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	fmt.Println("successfully created the event!")
	return nil
}

func deleteEvent(currentTheme *ui.Theme) error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("error obtaining config directory: %w", err)
	}

	fullConfig, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	if len(fullConfig.Events) == 0 {
		fmt.Println("no events found to delete.")
		return nil
	}

	var options []huh.Option[int]
	for idx, evt := range fullConfig.Events {
		label := fmt.Sprintf("%s (%s)", evt.Title, evt.Date)
		options = append(options, huh.NewOption(label, idx))
	}

	var selectedIdx int

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("select an event to delete:").
				Options(options...).
				Value(&selectedIdx),
		),
	).WithTheme(currentTheme.HuhTheme)

	if err := form.Run(); err != nil {
		return err
	}

	if err := DeleteEventAtIndex(fullConfig, selectedIdx); err != nil {
		return err
	}

	if err := config.SaveConfig(configPath, fullConfig); err != nil {
		return fmt.Errorf("error saving config after deletion: %w", err)
	}

	fmt.Println("successfully deleted the event!")
	return nil
}

func EventsWizard(currentTheme *ui.Theme) error {
	var action string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("event management").
				Description("choose an action to manage your events:").
				Options(
					huh.NewOption("add a new event", "add"),
					huh.NewOption("delete an existing event", "delete"),
					huh.NewOption("go back", "back"),
				).
				Value(&action),
		),
	).WithTheme(currentTheme.HuhTheme)

	if err := form.Run(); err != nil {
		return err
	}

	switch action {
	case "add":
		if err := newEvent(currentTheme); err != nil {
			return fmt.Errorf("error creating new event: %w", err)
		}
	case "delete":
		if err := deleteEvent(currentTheme); err != nil {
			return fmt.Errorf("error deleting event: %w", err)
		}
	case "back":
		return nil
	}

	return nil
}
