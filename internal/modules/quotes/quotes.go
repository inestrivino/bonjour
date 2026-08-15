package quotes

// The quotes package manages the obtaining, caching and rendering of quotes for the application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/inestrivino/bonjour/internal/ui"
)

const (
	quoteAPIURL = "https://zenquotes.io/api/quotes"
)

// Quote represents a single quote structure
type Quote struct {
	Text   string `json:"q"` // ZenQuotes uses 'q' for the quote body
	Author string `json:"a"` // ZenQuotes uses 'a' for the author name
}

// LocalCache is the cache system for quotes
type LocalCache struct {
	LastDate string  `json:"last_date"` // Format: YYYY-MM-DD
	Current  Quote   `json:"current"`   // Today's quote
	Pool     []Quote `json:"pool"`      // Offline backup pool
}

func getCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	appCacheDir := filepath.Join(cacheDir, "bonjour")
	if err := os.MkdirAll(appCacheDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appCacheDir, "quotes_pool.json"), nil
}

func loadCache() (*LocalCache, error) {
	cachePath, err := getCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cache LocalCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

func saveCache(cache *LocalCache) {
	cachePath, err := getCachePath()
	if err != nil {
		return
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return
	}

	_ = os.WriteFile(cachePath, data, 0644)
}

// fetchNewBatch fetches 50 curated quotes from ZenQuotes API
func fetchNewBatch() ([]Quote, error) {
	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get(quoteAPIURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var quotes []Quote
	if err := json.NewDecoder(resp.Body).Decode(&quotes); err != nil {
		return nil, err
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("empty quote pool")
	}

	return quotes, nil
}

// GetDailyQuote guarantees instant startup by reusing today's quote or picking from the pool.
func GetDailyQuote() (*Quote, error) {
	today := time.Now().Format("2006-01-02")
	cache, err := loadCache()

	// Return today's quote if already fetched today
	if err == nil && cache.LastDate == today && cache.Current.Text != "" {
		return &cache.Current, nil
	}

	if cache == nil {
		cache = &LocalCache{}
	}

	// Refill the pool if exhausted
	if len(cache.Pool) == 0 {
		pool, fetchErr := fetchNewBatch()
		if fetchErr != nil {
			// Fallback: reuse previous quote if offline
			if cache.Current.Text != "" {
				return &cache.Current, nil
			}
			return nil, fmt.Errorf("offline and no cached quotes available: %w", fetchErr)
		}
		cache.Pool = pool
	}

	// Take the first quote from the batch for today and save remaining pool
	cache.Current = cache.Pool[0]
	cache.Pool = cache.Pool[1:]
	cache.LastDate = today

	saveCache(cache)

	return &cache.Current, nil
}

// RenderQuote renders the quote using the app's current theme.
func RenderQuote(theme *ui.Theme, miniRender bool) string {
	q, err := GetDailyQuote()
	if err != nil {
		if miniRender {
			return theme.ErrorText.Render("Quote Unavailable")
		}
		content := fmt.Sprintf("%s\n\n%s",
			theme.ErrorText.Render("Quote Unavailable"),
			theme.Subtitle.Render(err.Error()),
		)
		return theme.Card.BorderForeground(theme.Muted).Render(content)
	}

	if miniRender {
		return fmt.Sprintf("%s %s",
			theme.Body.Italic(true).Render(fmt.Sprintf("“%s”", q.Text)),
			theme.Subtitle.Render(fmt.Sprintf("— %s", q.Author)),
		)
	}

	content := fmt.Sprintf("%s\n\n%s",
		theme.Body.Italic(true).Render(fmt.Sprintf("“%s”", q.Text)),
		theme.Title.Align(lipgloss.Right).Render(fmt.Sprintf("— %s", q.Author)),
	)

	return theme.Card.Render(content)
}
