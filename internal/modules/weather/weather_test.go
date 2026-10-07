package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/inestrivino/bonjour/internal/ui"
)

// Mock terminal width for rendering test
var globalWidth = 80

// Helper to construct a mock UI Theme for testing LipGloss output
func mockTheme() *ui.Theme {
	return &ui.Theme{
		Title:     lipgloss.NewStyle(),
		Subtitle:  lipgloss.NewStyle(),
		Body:      lipgloss.NewStyle(),
		Banner:    lipgloss.NewStyle(),
		Muted:     lipgloss.Color("240"),
		Secondary: lipgloss.Color("205"),
		ErrorText: lipgloss.NewStyle(),
		Card:      lipgloss.NewStyle(),
	}
}

// Helper to isolate cache directory cross-platform
func setupTestEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	t.Setenv("XDG_CACHE_HOME", tmpDir) // Linux / BSD
	t.Setenv("HOME", tmpDir)           // macOS
	t.Setenv("LOCALAPPDATA", tmpDir)   // Windows

	return tmpDir
}

// GEOCODING TESTS

// Test NormalizeCountryCode inputs, ISO codes, and aliases
func TestNormalizeCountryCode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"   ", ""},
		{"es", "ES"},
		{"US", "US"},
		{"Spain", "ES"},
		{"españa", "ES"},
		{"United Kingdom", "GB"},
		{"uk", "GB"},
		{"France", "FR"},
		{"Deutschland", "DE"},
		{"unknown country name XYZ", ""},
	}

	for _, tt := range tests {
		t.Run("input_"+tt.input, func(t *testing.T) {
			got := NormalizeCountryCode(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeCountryCode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Test ResolveLocation API matching logic and error states
func TestResolveLocation(t *testing.T) {
	t.Run("Empty city returns nil without network call", func(t *testing.T) {
		res, err := ResolveLocation("", "Spain", "")
		if err != nil || res != nil {
			t.Errorf("expected (nil, nil) for empty city, got (%v, %v)", res, err)
		}
	})

	t.Run("Successful top match resolution", func(t *testing.T) {
		mockResp := geocodingResponse{
			Results: []LocationResult{
				{Name: "Paris", Latitude: 48.8566, Longitude: 2.3522, Country: "France", Admin1: "Île-de-France"},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("name") != "Paris" {
				t.Errorf("expected query param name=Paris, got %s", r.URL.Query().Get("name"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockResp)
		}))
		defer server.Close()

		origURL := geocodingBaseURL
		geocodingBaseURL = server.URL
		defer func() { geocodingBaseURL = origURL }()

		res, err := ResolveLocation("Paris", "France", "")
		if err != nil {
			t.Fatalf("ResolveLocation failed: %v", err)
		}

		if res.Name != "Paris" || res.Latitude != 48.8566 {
			t.Errorf("unexpected location result: %+v", res)
		}
	})

	t.Run("Filter by Admin1 region match", func(t *testing.T) {
		mockResp := geocodingResponse{
			Results: []LocationResult{
				{Name: "Springfield", Admin1: "Illinois", Latitude: 39.7817},
				{Name: "Springfield", Admin1: "Missouri", Latitude: 37.2089},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(mockResp)
		}))
		defer server.Close()

		origURL := geocodingBaseURL
		geocodingBaseURL = server.URL
		defer func() { geocodingBaseURL = origURL }()

		res, err := ResolveLocation("Springfield", "US", "Missouri")
		if err != nil {
			t.Fatalf("ResolveLocation failed: %v", err)
		}

		if res.Admin1 != "Missouri" || res.Latitude != 37.2089 {
			t.Errorf("expected Missouri match, got %+v", res)
		}
	})

	t.Run("Admin1 region mismatch error", func(t *testing.T) {
		mockResp := geocodingResponse{
			Results: []LocationResult{
				{Name: "Madrid", Admin1: "Comunidad de Madrid"},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(mockResp)
		}))
		defer server.Close()

		origURL := geocodingBaseURL
		geocodingBaseURL = server.URL
		defer func() { geocodingBaseURL = origURL }()

		_, err := ResolveLocation("Madrid", "ES", "NonExistentRegion")
		if err == nil || !strings.Contains(err.Error(), "failed to match region/state") {
			t.Errorf("expected admin mismatch error, got: %v", err)
		}
	})

	t.Run("No matching city results", func(t *testing.T) {
		mockResp := geocodingResponse{Results: []LocationResult{}}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(mockResp)
		}))
		defer server.Close()

		origURL := geocodingBaseURL
		geocodingBaseURL = server.URL
		defer func() { geocodingBaseURL = origURL }()

		_, err := ResolveLocation("NonExistentCity12345", "", "")
		if err == nil || !strings.Contains(err.Error(), "no matching location found") {
			t.Errorf("expected no location found error, got: %v", err)
		}
	})
}

// WEATHER DATA & ASCII TESTS

// Test ASCII Art Decision Branches based on conditions
func TestGetASCIIArt(t *testing.T) {
	tests := []struct {
		name     string
		res      *WeatherResult
		contains string
	}{
		{
			name:     "Rain condition (precip > 40%)",
			res:      &WeatherResult{PrecProb: 50, CloudCover: 80, IsDay: true},
			contains: "/ / /",
		},
		{
			name:     "Cloudy condition (clouds > 60%)",
			res:      &WeatherResult{PrecProb: 10, CloudCover: 75, IsDay: true},
			contains: "_(  )_( )_",
		},
		{
			name:     "Night condition (IsDay = false)",
			res:      &WeatherResult{PrecProb: 0, CloudCover: 10, IsDay: false},
			contains: ".-.",
		},
		{
			name:     "Sunny Day condition",
			res:      &WeatherResult{PrecProb: 0, CloudCover: 10, IsDay: true},
			contains: "- * *O * * -",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			art := getASCIIArt(tt.res)
			if !strings.Contains(art, tt.contains) {
				t.Errorf("getASCIIArt() expected to contain %q, got: %s", tt.contains, art)
			}
		})
	}
}

func TestIsExpired(t *testing.T) {
	t.Run("Not expired (recent timestamp)", func(t *testing.T) {
		// A timestamp from 30 minutes ago should NOT be expired
		recentTime := time.Now().Add(-30 * time.Minute)

		if isExpired(recentTime) {
			t.Error("expected recent timestamp to not be expired, got true")
		}
	})

	t.Run("Expired (older than 60 minutes)", func(t *testing.T) {
		// A timestamp from 65 minutes ago SHOULD be expired
		oldTime := time.Now().Add(-65 * time.Minute)

		if !isExpired(oldTime) {
			t.Error("expected old timestamp to be expired, got false")
		}
	})

	t.Run("Edge case (exactly 60 minutes ago)", func(t *testing.T) {
		// Exactly 60 minutes ago (expired)
		exactTime := time.Now().Add(-60 * time.Minute)

		if !isExpired(exactTime) {
			t.Error("expected timestamp at exactly 60 minutes to be expired")
		}
	})
}

func TestCacheLoadAndSave(t *testing.T) {
	_ = setupTestEnv(t)

	// Verify load on missing file returns error
	_, err := loadWeatherCache()
	if err == nil {
		t.Error("expected error loading non-existent cache file, got nil")
	}

	sampleCache := &CachedWeather{
		Timestamp: time.Now(),
		Data: WeatherResult{
			IsDay:       true,
			MaxTemp:     26.0,
			MinTemp:     15.0,
			PrecProb:    10,
			CurrentTemp: 22.5,
			CloudCover:  20,
		},
	}

	// Save cache to disk
	saveWeatherCache(sampleCache)

	// Reload cache and verify fields
	loaded, err := loadWeatherCache()
	if err != nil {
		t.Fatalf("loadCache() returned error after saving: %v", err)
	}

	if !loaded.Data.IsDay {
		t.Errorf("expected IsDay to be true, got false")
	}
	if loaded.Data.CurrentTemp != 22.5 {
		t.Errorf("expected CurrentTemp 22.5, got %f", loaded.Data.CurrentTemp)
	}
	if loaded.Data.MaxTemp != 26.0 || loaded.Data.MinTemp != 15.0 {
		t.Errorf("unexpected max/min temperatures: max=%f, min=%f", loaded.Data.MaxTemp, loaded.Data.MinTemp)
	}
}

// Test fetchWeatherData decoding and error conditions
func TestFetchWeatherData_MockHTTP(t *testing.T) {
	t.Run("Successful API Fetch", func(t *testing.T) {
		_ = setupTestEnv(t)

		var mockResp APIResponse
		mockResp.Current.IsDay = 1
		mockResp.Current.Temperature2m = 22.5
		mockResp.Current.CloudCover = 20
		mockResp.Current.PrecipitationProbability = 10
		mockResp.Daily.Temperature2mMax = []float64{26.0}
		mockResp.Daily.Temperature2mMin = []float64{15.0}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockResp)
		}))
		defer server.Close()

		origURL := weatherBaseURL
		weatherBaseURL = server.URL
		defer func() { weatherBaseURL = origURL }()

		res, err := fetchWeatherData(40.4168, -3.7038)
		if err != nil {
			t.Fatalf("fetchWeatherData failed: %v", err)
		}

		if !res.IsDay || res.CurrentTemp != 22.5 || res.MaxTemp != 26.0 || res.MinTemp != 15.0 {
			t.Errorf("unexpected weather result: %+v", res)
		}
	})

	t.Run("HTTP Server Error", func(t *testing.T) {
		_ = setupTestEnv(t)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		origURL := weatherBaseURL
		weatherBaseURL = server.URL
		defer func() { weatherBaseURL = origURL }()

		_, err := fetchWeatherData(0, 0)
		if err == nil || !strings.Contains(err.Error(), "non-valid http response") {
			t.Errorf("expected HTTP error, got: %v", err)
		}
	})

	t.Run("HTTP Error with Cache Fallback", func(t *testing.T) {
		_ = setupTestEnv(t)
		cachedData := &CachedWeather{
			Timestamp: time.Now(),
			Data: WeatherResult{
				IsDay:       true,
				MaxTemp:     25.0,
				MinTemp:     14.0,
				PrecProb:    5,
				CurrentTemp: 20.0,
				CloudCover:  15,
			},
		}
		saveWeatherCache(cachedData)

		//Point weatherBaseURL to the address of a closed server to trigger http error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		serverURL := server.URL
		server.Close()

		origURL := weatherBaseURL
		weatherBaseURL = serverURL
		defer func() { weatherBaseURL = origURL }()

		// because fetchweatherdata gets a http error, it loads the saved cache
		result, err := fetchWeatherData(40.4168, -3.7038)
		if err != nil {
			t.Fatalf("expected fallback to succeed, got error: %v", err)
		}

		if result.CurrentTemp != 20.0 {
			t.Errorf("expected fallback current temp 20.0, got %f", result.CurrentTemp)
		}
	})
}

// Test RenderWeatherData Output Layouts (Full & Mini View)
func TestRenderWeatherData(t *testing.T) {
	theme := mockTheme()

	var mockResp APIResponse
	mockResp.Current.IsDay = 1
	mockResp.Current.Temperature2m = 18.0
	mockResp.Current.CloudCover = 10
	mockResp.Current.PrecipitationProbability = 0
	mockResp.Daily.Temperature2mMax = []float64{22.0}
	mockResp.Daily.Temperature2mMin = []float64{12.0}

	t.Run("Standard Card Rendering", func(t *testing.T) {
		_ = setupTestEnv(t)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockResp)
		}))
		defer server.Close()

		origURL := weatherBaseURL
		weatherBaseURL = server.URL
		defer func() { weatherBaseURL = origURL }()

		out := RenderWeatherData(40.4168, -3.7038, "Madrid", theme, false, globalWidth)

		if !strings.Contains(out, "Madrid") {
			t.Error("expected output to contain city name 'Madrid'")
		}
		if !strings.Contains(out, "18.0°C") {
			t.Error("expected output to contain current temperature '18.0°C'")
		}
	})

	t.Run("Mini View Rendering", func(t *testing.T) {
		_ = setupTestEnv(t)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockResp)
		}))
		defer server.Close()

		origURL := weatherBaseURL
		weatherBaseURL = server.URL
		defer func() { weatherBaseURL = origURL }()

		out := RenderWeatherData(40.4168, -3.7038, "Madrid", theme, true, globalWidth)

		if !strings.Contains(out, "Madrid: 18.0°C (H: 22.0°C / L: 12.0°C)") {
			t.Errorf("unexpected mini render output: %s", out)
		}
	})

	t.Run("Error Fallback Rendering", func(t *testing.T) {
		_ = setupTestEnv(t)

		// fake server to receive http error
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer errorServer.Close()

		origURL := weatherBaseURL
		weatherBaseURL = errorServer.URL
		defer func() { weatherBaseURL = origURL }()

		out := RenderWeatherData(0, 0, "Unknown", theme, false, globalWidth)
		if !strings.Contains(out, "Weather Unavailable") {
			t.Errorf("expected error fallback title 'Weather Unavailable', got: %s", out)
		}
	})
}
