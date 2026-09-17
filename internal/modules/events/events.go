package events

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

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

	if newEvent != nil {
		fullConfig.Events = append(fullConfig.Events, *newEvent)
	}

	if err := config.SaveConfig(configPath, fullConfig); err != nil {
		return fmt.Errorf("error saving updated configuration: %w", err)
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

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	eventDate := time.Date(eventTime.Year(), eventTime.Month(), eventTime.Day(), 0, 0, 0, 0, now.Location())

	days := int(eventDate.Sub(today).Hours() / 24)
	return days, nil
}

// RenderEvents displays up to the 5 closest upcoming events on the chosen theme.
func RenderEvents(theme *ui.Theme, miniRender bool, width int) string {
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
		return theme.Card.Width(width).Render(theme.Subtitle.Render("No upcoming events scheduled."))
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
			return theme.Subtitle.Render("No active event reminders for today")
		}
		return theme.Card.Width(width).Render(theme.Subtitle.Render("No active event reminders for today"))
	}

	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].days < upcoming[j].days
	})

	if miniRender {
		item := upcoming[0]
		daysText := fmt.Sprintf("%dd left", item.days)
		if item.days == 0 {
			daysText = "Today!"
		}
		return fmt.Sprintf("%s: %s", theme.Title.Render(item.event.Title), theme.Body.Render(daysText))
	}

	if len(upcoming) > 5 {
		upcoming = upcoming[:5]
	}

	innerWidth := width - 4
	if innerWidth < 10 {
		innerWidth = 10
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
		row := lipgloss.JoinHorizontal(lipgloss.Center, titleText, "   ", daysText)
		rows = append(rows, lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(row))
	}

	header := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(theme.Title.Render("Upcoming Events"))
	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	content := lipgloss.JoinVertical(lipgloss.Left, header, "", body)

	return theme.Card.Width(width).Render(content)
}

// Helper function to render errors in a uniform way
func renderErrorCard(theme *ui.Theme, title string, detail string) string {
	content := fmt.Sprintf("%s\n\n%s",
		theme.ErrorText.Render(title),
		theme.Subtitle.Render(detail),
	)
	return theme.Card.BorderForeground(theme.Muted).Render(content)
}

// ValidateEventTitle ensures that the event's title is not empty or invalid
func ValidateEventTitle(str string) error {
	if strings.TrimSpace(str) == "" {
		return errors.New("title cannot be empty")
	}
	return nil
}

// ValidateEventDate ensures that the event's date is correctly formatted
func ValidateEventDate(str string) error {
	_, err := time.Parse("2006-01-02", strings.TrimSpace(str))
	if err != nil {
		return errors.New("invalid date format; please use YYYY-MM-DD")
	}
	return nil
}

// ValidateWarningDate ensures the date established as start of warning is valid
func ValidateWarningDate(warningStr, eventDateStr string, now time.Time) error {
	warningStr = strings.TrimSpace(warningStr)
	var wTime time.Time
	var err error

	if warningStr == "" {
		todayStr := now.Format("2006-01-02")
		wTime, _ = time.Parse("2006-01-02", todayStr)
	} else {
		wTime, err = time.Parse("2006-01-02", warningStr)
		if err != nil {
			return errors.New("invalid date format; please use YYYY-MM-DD")
		}
	}

	if eTime, err := time.Parse("2006-01-02", strings.TrimSpace(eventDateStr)); err == nil {
		if wTime.After(eTime) {
			return errors.New("warning date cannot be after the event date")
		}
	}
	return nil
}

// CreateEventObject returns a new Event object
func CreateEventObject(title, dateStr, warningStart string) *config.EventConfig {
	return &config.EventConfig{
		Title:        strings.TrimSpace(title),
		Date:         strings.TrimSpace(dateStr),
		WarningStart: strings.TrimSpace(warningStart),
	}
}

// DeleteEventAtIndex handles the deletion of an Event object from the main Config object
func DeleteEventAtIndex(fullConfig *config.Config, index int) error {
	if index < 0 || index >= len(fullConfig.Events) {
		return errors.New("event index out of bounds")
	}
	fullConfig.Events = append(fullConfig.Events[:index], fullConfig.Events[index+1:]...)
	return nil
}
