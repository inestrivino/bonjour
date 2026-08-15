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
	// Fetch 30 quotes in one go to minimize network requests
	quoteAPIURL = "https://dummyjson.com/quotes?limit=30"
)

// Quote is the structure of a quote from the API
type Quote struct {
	Text   string `json:"quote"`
	Author string `json:"author"`
}

// APIResponse is the list of quotes received from the API call
type APIResponse struct {
	Quotes []Quote `json:"quotes"`
}

// LocalCache is the cache system for quotes
type LocalCache struct {
	LastDate string  `json:"last_date"` // Format: YYYY-MM-DD
	Current  Quote   `json:"current"`   // Today's quote
	Pool     []Quote `json:"pool"`      // Offline backup pool
}

// getCachePath gets or creates a directory in the user's cache directory for the questions to be stored.
// It returns the path to the file of cached questions, it may also return an error
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

// loadCache obtains the LocalCache object stored in the cache directory. It may also return an error.
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

// saveCache takes in a LocalCache object and inserts it into the cache file
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

// fetchNewBatch fetches 30 quotes from the API to populate the pool, it may also return an error
func fetchNewBatch() ([]Quote, error) {
	//connection closed after 3 seconds without anser
	client := &http.Client{Timeout: 3 * time.Second}
	//we obtain the response and possibly an error
	resp, err := client.Get(quoteAPIURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	//if the information was received successfully we decode it and send it back as an array of Quote object
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Quotes) == 0 {
		return nil, fmt.Errorf("empty quote pool")
	}

	return apiResp.Quotes, nil
}

// GetDailyQuote guarantees instant startup by reusing today's quote or picking from the pool. It returns a Quote object, and may return an error
func GetDailyQuote() (*Quote, error) {
	today := time.Now().Format("2006-01-02")
	cache, err := loadCache()

	// Return today's quote
	if err == nil && cache.LastDate == today && cache.Current.Text != "" {
		return &cache.Current, nil
	}

	// If there is no "today's quote", then we preare caching system
	if cache == nil {
		cache = &LocalCache{}
	}

	// We call the server for a new batch of quotes
	if len(cache.Pool) == 0 {
		pool, fetchErr := fetchNewBatch()
		if fetchErr != nil {
			// Fallback: If network fails and we have a previous current quote, reuse it
			if cache.Current.Text != "" {
				return &cache.Current, nil
			}
			return nil, fmt.Errorf("offline and no cached quotes available: %w", fetchErr)
		}
		cache.Pool = pool
	}

	// After obtaining the new pool of questions, we take one to use for today and leave the rest
	cache.Current = cache.Pool[0]
	cache.Pool = cache.Pool[1:]
	cache.LastDate = today

	// Save state to disk
	saveCache(cache)

	return &cache.Current, nil
}

// RenderQuote takes in the app's theme and renders the quote based on that. It returns the string text to be printed.
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
		// Single-line compact view
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
