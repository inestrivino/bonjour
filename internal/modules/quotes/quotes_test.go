package quotes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/inestrivino/bonjour/internal/ui"
)

// Helper to isolate cache directory cross-platform
func setupTestEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	t.Setenv("XDG_CACHE_HOME", tmpDir) // Linux / BSD
	t.Setenv("HOME", tmpDir)           // macOS
	t.Setenv("LOCALAPPDATA", tmpDir)   // Windows

	return tmpDir
}

// Helper to construct a mock UI Theme for testing LipGloss rendering
func mockTheme() *ui.Theme {
	return &ui.Theme{
		Title:     lipgloss.NewStyle(),
		Subtitle:  lipgloss.NewStyle(),
		Body:      lipgloss.NewStyle(),
		Muted:     lipgloss.Color("240"),
		ErrorText: lipgloss.NewStyle(),
		Card:      lipgloss.NewStyle(),
	}
}

// Test getCachePath cross-platform folder creation
func TestGetCachePath(t *testing.T) {
	tmpDir := setupTestEnv(t)

	cachePath, err := getCachePath()
	if err != nil {
		t.Fatalf("getCachePath() returned unexpected error: %v", err)
	}

	if !strings.HasPrefix(cachePath, tmpDir) {
		t.Errorf("expected cache path to be inside temp dir %s, got: %s", tmpDir, cachePath)
	}

	if !strings.HasSuffix(cachePath, filepath.Join("bonjour", "quotes_pool.json")) {
		t.Errorf("expected path ending with 'bonjour/quotes_pool.json', got: %s", cachePath)
	}

	// Verify directory was created on disk
	dir := filepath.Dir(cachePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("getCachePath() failed to create cache directory on disk")
	}
}

// Test Cache Loading and Saving
func TestCacheLoadAndSave(t *testing.T) {
	_ = setupTestEnv(t)

	// Verify load on missing file returns error
	_, err := loadCache()
	if err == nil {
		t.Error("expected error loading non-existent cache file, got nil")
	}

	sampleCache := &LocalCache{
		LastDate: "2026-08-15",
		Current: Quote{
			Text:   "The journey of a thousand miles begins with a single step.",
			Author: "Lao Tzu",
		},
		Pool: []Quote{
			{Text: "Be yourself; everyone else is already taken.", Author: "Oscar Wilde"},
		},
	}

	// Save cache to disk
	saveCache(sampleCache)

	// Reload cache and verify fields
	loaded, err := loadCache()
	if err != nil {
		t.Fatalf("loadCache() returned error after saving: %v", err)
	}

	if loaded.LastDate != "2026-08-15" {
		t.Errorf("expected LastDate '2026-08-15', got '%s'", loaded.LastDate)
	}
	if loaded.Current.Author != "Lao Tzu" {
		t.Errorf("expected Author 'Lao Tzu', got '%s'", loaded.Current.Author)
	}
	if len(loaded.Pool) != 1 || loaded.Pool[0].Author != "Oscar Wilde" {
		t.Errorf("unexpected pool content: %+v", loaded.Pool)
	}
}

// Test HTTP Fetch Batching using httptest.Server
func TestFetchNewBatch_MockHTTP(t *testing.T) {
	t.Run("Successful Fetch", func(t *testing.T) {
		mockQuotes := []Quote{
			{Text: "Test Quote 1", Author: "Author 1"},
			{Text: "Test Quote 2", Author: "Author 2"},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockQuotes)
		}))
		defer server.Close()

		// Temporarily override API URL
		origURL := quoteAPIURL
		quoteAPIURL = server.URL
		defer func() { quoteAPIURL = origURL }()

		quotes, err := fetchNewBatch()
		if err != nil {
			t.Fatalf("fetchNewBatch() failed: %v", err)
		}

		if len(quotes) != 2 {
			t.Fatalf("expected 2 quotes, got %d", len(quotes))
		}
		if quotes[0].Text != "Test Quote 1" {
			t.Errorf("expected 'Test Quote 1', got '%s'", quotes[0].Text)
		}
	})

	t.Run("HTTP Server Error 500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		origURL := quoteAPIURL
		quoteAPIURL = server.URL
		defer func() { quoteAPIURL = origURL }()

		_, err := fetchNewBatch()
		if err == nil {
			t.Error("expected error for HTTP 500, got nil")
		}
	})

	t.Run("Empty Response Array Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]Quote{})
		}))
		defer server.Close()

		origURL := quoteAPIURL
		quoteAPIURL = server.URL
		defer func() { quoteAPIURL = origURL }()

		_, err := fetchNewBatch()
		if err == nil || !strings.Contains(err.Error(), "empty quote pool") {
			t.Errorf("expected 'empty quote pool' error, got: %v", err)
		}
	})
}

// Test GetDailyQuote Cache, Refill, and Offline Fallback Flow
func TestGetDailyQuote(t *testing.T) {
	_ = setupTestEnv(t)
	today := time.Now().Format("2006-01-02")

	t.Run("Reuse Cache from Today", func(t *testing.T) {
		cache := &LocalCache{
			LastDate: today,
			Current: Quote{
				Text:   "Cached Today",
				Author: "Test Author",
			},
		}
		saveCache(cache)

		quote, err := GetDailyQuote()
		if err != nil {
			t.Fatalf("GetDailyQuote() failed: %v", err)
		}

		if quote.Text != "Cached Today" {
			t.Errorf("expected 'Cached Today', got '%s'", quote.Text)
		}
	})

	t.Run("Pool Exhaustion Refill from Network", func(t *testing.T) {
		// Set cache date to yesterday with empty pool
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		cache := &LocalCache{
			LastDate: yesterday,
			Current:  Quote{Text: "Old Quote", Author: "Old Author"},
			Pool:     []Quote{},
		}
		saveCache(cache)

		mockQuotes := []Quote{
			{Text: "Fresh Quote 1", Author: "New Author 1"},
			{Text: "Fresh Quote 2", Author: "New Author 2"},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(mockQuotes)
		}))
		defer server.Close()

		origURL := quoteAPIURL
		quoteAPIURL = server.URL
		defer func() { quoteAPIURL = origURL }()

		quote, err := GetDailyQuote()
		if err != nil {
			t.Fatalf("GetDailyQuote() failed: %v", err)
		}

		if quote.Text != "Fresh Quote 1" {
			t.Errorf("expected 'Fresh Quote 1', got '%s'", quote.Text)
		}

		// Verify pool was popped and saved to disk
		updatedCache, _ := loadCache()
		if len(updatedCache.Pool) != 1 {
			t.Errorf("expected 1 remaining quote in pool, got %d", len(updatedCache.Pool))
		}
		if updatedCache.LastDate != today {
			t.Errorf("expected LastDate updated to today (%s), got %s", today, updatedCache.LastDate)
		}
	})

	t.Run("Offline Fallback Reuses Previous Quote", func(t *testing.T) {
		_ = setupTestEnv(t)
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

		cache := &LocalCache{
			LastDate: yesterday,
			Current:  Quote{Text: "Fallback Stale Quote", Author: "Fallback Author"},
			Pool:     []Quote{},
		}
		saveCache(cache)

		// Point to broken HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		origURL := quoteAPIURL
		quoteAPIURL = server.URL
		defer func() { quoteAPIURL = origURL }()

		quote, err := GetDailyQuote()
		if err != nil {
			t.Fatalf("expected fallback quote when offline, got error: %v", err)
		}

		if quote.Text != "Fallback Stale Quote" {
			t.Errorf("expected 'Fallback Stale Quote', got '%s'", quote.Text)
		}
	})

	t.Run("Offline and Completely Empty Cache Error", func(t *testing.T) {
		_ = setupTestEnv(t) // Empty cache environment

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		origURL := quoteAPIURL
		quoteAPIURL = server.URL
		defer func() { quoteAPIURL = origURL }()

		_, err := GetDailyQuote()
		if err == nil {
			t.Error("expected error when offline with no cached quote, got nil")
		}
	})
}

// Test RenderQuote output formatting and mini vs standard modes
func TestRenderQuote(t *testing.T) {
	_ = setupTestEnv(t)
	theme := mockTheme()

	today := time.Now().Format("2006-01-02")
	cache := &LocalCache{
		LastDate: today,
		Current: Quote{
			Text:   "Stay hungry, stay foolish.",
			Author: "Steve Jobs",
		},
	}
	saveCache(cache)

	t.Run("Standard View Rendering", func(t *testing.T) {
		out := RenderQuote(theme, false)

		if !strings.Contains(out, "“Stay hungry, stay foolish.”") {
			t.Errorf("expected formatted quote text in output, got: %s", out)
		}
		if !strings.Contains(out, "— Steve Jobs") {
			t.Errorf("expected formatted author in output, got: %s", out)
		}
	})

	t.Run("Mini View Rendering", func(t *testing.T) {
		out := RenderQuote(theme, true)

		if !strings.Contains(out, "“Stay hungry, stay foolish.”") || !strings.Contains(out, "— Steve Jobs") {
			t.Errorf("unexpected mini view rendering: %s", out)
		}
	})

	t.Run("Error Card Rendering when Quote Fails", func(t *testing.T) {
		_ = setupTestEnv(t) // Wipe cache to trigger failure

		// Point API to broken server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		origURL := quoteAPIURL
		quoteAPIURL = server.URL
		defer func() { quoteAPIURL = origURL }()

		stdOut := RenderQuote(theme, false)
		if !strings.Contains(stdOut, "Quote Unavailable") {
			t.Errorf("expected error title 'Quote Unavailable', got: %s", stdOut)
		}

		miniOut := RenderQuote(theme, true)
		if !strings.Contains(miniOut, "Quote Unavailable") {
			t.Errorf("expected mini error string 'Quote Unavailable', got: %s", miniOut)
		}
	})
}
