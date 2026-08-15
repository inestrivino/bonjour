package events

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/inestrivino/bonjour/internal/config"
	"github.com/inestrivino/bonjour/internal/ui"
)

// saveEvent parses and saves a new event into the user config
func saveEvent(newEvent *config.EventConfig) error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("error obtaining config directory: %w", err)
	}

	fullConfig, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Append the new event to the existing events slice
	if newEvent != nil {
		fullConfig.Events = append(fullConfig.Events, *newEvent)
	}

	// Save the updated configuration back to disk
	if err := config.SaveConfig(configPath, fullConfig); err != nil {
		return fmt.Errorf("error saving updated configuration: %w", err)
	}

	return nil
}

// newEvent displays a form to the user from which it takes the data to create a new Event object, then saves it into the disk
func newEvent(currentTheme *ui.Theme) error {
	// Structure for the new event
	var (
		title        string
		dateStr      string
		warningStart string
	)

	// Form for user to fill and thus create the event.
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Event Title").
				Placeholder("Mom's Birthday").
				Value(&title).
				Validate(func(str string) error {
					if strings.TrimSpace(str) == "" {
						return errors.New("title cannot be empty")
					}
					return nil
				}),

			huh.NewInput().
				Title("Event Date").
				Description("Format: YYYY-MM-DD").
				Placeholder("2026-12-25").
				Value(&dateStr).
				Validate(func(str string) error {
					_, err := time.Parse("2006-01-02", strings.TrimSpace(str))
					if err != nil {
						return errors.New("invalid date format; please use YYYY-MM-DD")
					}
					return nil
				}),

			huh.NewInput().
				Title("Warning Start Date").
				Description("From which date forward do you want to be reminded? (YYYY-MM-DD). Leave empty to start today.").
				Placeholder("2026-12-18").
				Value(&warningStart).
				Validate(func(str string) error {
					str = strings.TrimSpace(str)
					var wTime time.Time
					var err error
					if str == "" {
						// If str is empty, default warning time to today
						todayStr := time.Now().Format("2006-01-02")
						wTime, _ = time.Parse("2006-01-02", todayStr)
					} else {
						wTime, err = time.Parse("2006-01-02", str)
						if err != nil {
							return errors.New("invalid date format; please use YYYY-MM-DD")
						}
					}

					if err != nil {
						return errors.New("invalid date format; please use YYYY-MM-DD")
					}

					// Verify warning start isn't after the event date
					if eTime, err := time.Parse("2006-01-02", strings.TrimSpace(dateStr)); err == nil {
						if wTime.After(eTime) {
							return errors.New("warning date cannot be after the event date")
						}
					}
					return nil
				}),
		),
	).WithTheme(currentTheme.HuhTheme)

	if err := form.Run(); err != nil {
		return err
	}

	// Turn the event variable into an EventConfig type
	event := &config.EventConfig{
		Title:        strings.TrimSpace(title),
		Date:         strings.TrimSpace(dateStr),
		WarningStart: strings.TrimSpace(warningStart),
	}

	// Now we save the EventConfig object
	if err := saveEvent(event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	return nil
}

// deleteEvent displays a form for the user to choose an event to delete from the config directory
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
		fmt.Println("No events found to delete.")
		return nil
	}

	// Build selectable options from existing events
	var options []huh.Option[int]
	for idx, evt := range fullConfig.Events {
		label := fmt.Sprintf("%s (%s)", evt.Title, evt.Date)
		options = append(options, huh.NewOption(label, idx))
	}

	var selectedIdx int

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select an event to delete:").
				Options(options...).
				Value(&selectedIdx),
		),
	).WithTheme(currentTheme.HuhTheme)

	if err := form.Run(); err != nil {
		return err
	}

	// Remove chosen element from slice preserving order
	fullConfig.Events = append(fullConfig.Events[:selectedIdx], fullConfig.Events[selectedIdx+1:]...)

	if err := config.SaveConfig(configPath, fullConfig); err != nil {
		return fmt.Errorf("error saving config after deletion: %w", err)
	}

	return nil
}

// daysLeft calculates, given an event's date, how many days are left from today until then.
func daysLeft(event *config.EventConfig) (int, error) {
	if event == nil {
		return 0, fmt.Errorf("event is nil")
	}

	eventTime, err := time.Parse("2006-01-02", event.Date)
	if err != nil {
		return 0, err
	}

	// Normalize today to midnight in local time for clean day-difference calculation
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Truncate event time to midnight as well
	eventDate := time.Date(eventTime.Year(), eventTime.Month(), eventTime.Day(), 0, 0, 0, 0, now.Location())

	days := int(eventDate.Sub(today).Hours() / 24)
	return days, nil
}

// RenderEvents displays up to the 5 closest upcoming events on the chosen theme.
func RenderEvents(theme *ui.Theme, miniRender bool) string {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return renderErrorCard(theme, "Events Unavailable", err.Error())
	}

	fullConfig, err := config.LoadConfig(configPath)
	if err != nil {
		return renderErrorCard(theme, "Events Unavailable", err.Error())
	}

	if len(fullConfig.Events) == 0 {
		if miniRender {
			return theme.Subtitle.Render("No events scheduled")
		}
		emptyContent := theme.Subtitle.Render("No upcoming events scheduled.")
		return theme.Card.Render(
			lipgloss.JoinVertical(lipgloss.Left, theme.Title.Render("Events"), "", emptyContent),
		)
	}

	type eventItem struct {
		event config.EventConfig
		days  int
	}

	now := time.Now()
	todayStr := now.Format("2006-01-02")

	var upcoming []eventItem
	for _, evt := range fullConfig.Events {
		if evt.WarningStart != "" && evt.WarningStart > todayStr {
			continue
		}

		days, err := daysLeft(&evt)
		if err != nil {
			continue
		}

		if days >= 0 {
			upcoming = append(upcoming, eventItem{event: evt, days: days})
		}
	}

	if len(upcoming) == 0 {
		if miniRender {
			return theme.Subtitle.Render("No active events today")
		}
		emptyContent := theme.Subtitle.Render("No active event reminders for today.")
		return theme.Card.Render(
			lipgloss.JoinVertical(lipgloss.Left, theme.Title.Render("Events"), "", emptyContent),
		)
	}

	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].days < upcoming[j].days
	})

	if miniRender {
		// Compact view with less events
		item := upcoming[0]
		daysText := fmt.Sprintf("%dd left", item.days)
		if item.days == 0 {
			daysText = "Today!"
		}
		return fmt.Sprintf("%s: %s",
			theme.Title.Render(item.event.Title),
			theme.Body.Render(daysText),
		)
	}

	// Standard view (top 5 events in full card)
	if len(upcoming) > 5 {
		upcoming = upcoming[:5]
	}

	var rows []string
	for _, item := range upcoming {
		var daysText string
		switch item.days {
		case 0:
			daysText = theme.ErrorText.Render("Today!")
		case 1:
			daysText = theme.Subtitle.Render("1 day left")
		default:
			daysText = theme.Subtitle.Render(fmt.Sprintf("%d days left", item.days))
		}

		titleText := theme.Body.Render(item.event.Title)
		dateText := lipgloss.NewStyle().Foreground(theme.Muted).Render(fmt.Sprintf("(%s)", item.event.Date))

		leftSide := fmt.Sprintf("%s %s", titleText, dateText)
		row := lipgloss.JoinHorizontal(lipgloss.Center, leftSide, "  —  ", daysText)
		rows = append(rows, row)
	}

	header := theme.Title.Render("Upcoming Events")
	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	content := lipgloss.JoinVertical(lipgloss.Left, header, "", body)

	return theme.Card.Render(content)
}

// Helper to format error states consistently with RenderQuote
func renderErrorCard(theme *ui.Theme, title string, detail string) string {
	content := fmt.Sprintf("%s\n\n%s",
		theme.ErrorText.Render(title),
		theme.Subtitle.Render(detail),
	)
	return theme.Card.BorderForeground(theme.Muted).Render(content)
}

// EventsWizard is the menu function for managing events.
func EventsWizard(currentTheme *ui.Theme) error {
	var action string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Event Management").
				Description("Choose an action to manage your events:").
				Options(
					huh.NewOption("Add a new event", "add"),
					huh.NewOption("Delete an existing event", "delete"),
					huh.NewOption("Go Back", "back"),
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
