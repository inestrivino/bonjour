package weather

// The weather package manages the functionalities regarding fetching and rendering weather data related to the user's location
// To relate a human input (city name + country and/or region) to a location that can be understood by the weather API, the geocoding.go file
// to an exact latitude and longitude which can then be sent to the weather API call.

// Weather data provided by Open-Meteo API (https://open-meteo.com/)
// Free for non-commercial use under CC BY 4.0.

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

// APIResponse is the entire structure of the response from the Open-Meteo API
type APIResponse struct {
	Current struct {
		IsDay                    int     `json:"is_day"`
		PrecipitationProbability int     `json:"precipitation_probability"`
		Temperature2m            float64 `json:"temperature_2m"`
		CloudCover               int     `json:"cloud_cover"`
	} `json:"current"`
	Daily struct {
		Temperature2mMax []float64 `json:"temperature_2m_max"`
		Temperature2mMin []float64 `json:"temperature_2m_min"`
	} `json:"daily"`
}

// WeatherResult is a simplified structure of the data given my the API for rendering purposes
type WeatherResult struct {
	IsDay       bool
	MaxTemp     float64
	MinTemp     float64
	PrecProb    int
	CurrentTemp float64
	CloudCover  int
}

// CachedWeather wraps the result alongside a timestamp to track expiration
type CachedWeather struct {
	Timestamp time.Time     `json:"timestamp"`
	Data      WeatherResult `json:"data"`
}

// getCachePath returns the file system path to the cache file for the application
func getWeatherCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	appCacheDir := filepath.Join(cacheDir, "bonjour")
	if err := os.MkdirAll(appCacheDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appCacheDir, "weather_cache.json"), nil
}

// loadQuotesCache takes the information within the cache file into a WeatherResult object
func loadWeatherCache() (*CachedWeather, error) {
	cachePath, err := getWeatherCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cached CachedWeather
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}
	return &cached, nil
}

func saveWeatherCache(cache *CachedWeather) {
	cachePath, err := getWeatherCachePath()
	if err != nil {
		return
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return
	}

	_ = os.WriteFile(cachePath, data, 0644)
}

var weatherBaseURL = "https://api.open-meteo.com/v1/forecast"

// fetchWeatherData takes in a latitude and longitude to perform an API call, it then saves it within the cache. It returns a WeatherResult type object and may return an error
func fetchWeatherData(lat, lon float64) (*WeatherResult, error) {
	// Full field needed for the API call
	weatherAPIURL := fmt.Sprintf(
		"%s?latitude=%.4f&longitude=%.4f&current=temperature_2m,precipitation_probability,cloud_cover,is_day&daily=temperature_2m_max,temperature_2m_min&forecast_days=1&timezone=auto",
		weatherBaseURL, lat, lon,
	)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(weatherAPIURL)

	// Fallback: If network fails, serve expired cache rather than crashing completely
	// It may fail due to having .Get return an error, or due to having it return a StatusCode different than OK
	if err != nil {

		if cachedResult, fallbackErr := loadWeatherCache(); fallbackErr == nil {
			return &cachedResult.Data, nil
		}
		return nil, fmt.Errorf("error during http request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if cachedResult, fallbackErr := loadWeatherCache(); fallbackErr == nil {
			return &cachedResult.Data, nil
		}
		return nil, fmt.Errorf("non-valid http response: %s", resp.Status)
	}

	// API response is valid, we decode it
	var apiData APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiData); err != nil {
		return nil, fmt.Errorf("error while decoding json: %w", err)
	}
	// and insert it into a weather result object
	result := WeatherResult{
		IsDay:       apiData.Current.IsDay == 1,
		CurrentTemp: apiData.Current.Temperature2m,
		CloudCover:  apiData.Current.CloudCover,
		PrecProb:    apiData.Current.PrecipitationProbability,
	}

	// Extraction of max and min temperatures today
	if len(apiData.Daily.Temperature2mMax) > 0 {
		result.MaxTemp = apiData.Daily.Temperature2mMax[0]
	}
	if len(apiData.Daily.Temperature2mMin) > 0 {
		result.MinTemp = apiData.Daily.Temperature2mMin[0]
	}

	// Save fresh data and timestamp to cache
	cached := CachedWeather{
		Timestamp: time.Now(),
		Data:      result,
	}
	saveWeatherCache(&cached)

	// Returns the object
	return &result, nil
}

// isExpired checks if the given timestamp is older than 60 minutes
func isExpired(timestamp time.Time) bool {
	return time.Since(timestamp) > 60*time.Minute
}

// GetWeather checks if it is time to fetch new weather data, otherwise it serves the cached data
func GetWeather(lat, lon float64) (*WeatherResult, error) {
	// We load the cached weather data
	weatherCache, err := loadWeatherCache()

	// If we receive an error (for example the cache document doesn't yet exist) there is no data, or it is expired, we fetch new data
	if err != nil || weatherCache == nil || isExpired(weatherCache.Timestamp) {
		// Fetch new data
		weather, err := fetchWeatherData(lat, lon)
		if err != nil {
			return nil, err
		}
		return weather, nil
	}

	return &weatherCache.Data, nil
}

// getAsciiArt receives a WeatherResult object, from which it determines the type of ascii art that should be rendered on the terminal. It returns a string
func getASCIIArt(w *WeatherResult) string {
	if w.PrecProb > 40 {
		return `
	      __   _
	    _(  )_( )_
	   (_   _    _)
	  / /(_) (__)
	 / / / / / /
	/ / / / / /
	 `
	}
	if w.CloudCover > 60 {
		return `
	   __   _
	 _(  )_( )_
	(_   _    _)
	  (_) (__)
	`
	}
	if !w.IsDay {
		return `
		   *  .-.
		     (   )
		      ` + "`" + `-'`
	}
	return `
	     \ | /
	      \*/
	  - * *O * * -
	      /*\
	     / | \
	`
}

// RenderWeatherData takes in the latitude and longitude, as well as a theme object, calls the fetch function and uses the theme information to render the result accordingly
func RenderWeatherData(lat, lon float64, city string, theme *ui.Theme, miniRender bool, width int) string {
	data, err := GetWeather(lat, lon)
	if err != nil {
		if miniRender {
			return theme.ErrorText.Render(fmt.Sprintf("Weather: %s", err.Error()))
		}
		return theme.Card.Width(width).Render(theme.ErrorText.Render("Weather Unavailable: ") + err.Error())
	}

	if miniRender {
		return fmt.Sprintf("%s: %.1f°C (H: %.1f°C / L: %.1f°C)",
			theme.Title.Render(city), data.CurrentTemp, data.MaxTemp, data.MinTemp)
	}

	innerWidth := width - 4
	if innerWidth < 10 {
		innerWidth = 10
	}

	header := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
		theme.Title.Render(fmt.Sprintf("%s in %s", time.Now().Format("02/01/2006"), city)),
	)
	asciiArt := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
		theme.Banner.Foreground(theme.Secondary).Render(getASCIIArt(data)),
	)

	rawMetrics := fmt.Sprintf(
		"%s %.1f°C\n%s %.1f°C / %.1f°C\n%s %d%%\n%s %d%%",
		theme.Subtitle.Render("Current:       "), data.CurrentTemp,
		theme.Subtitle.Render("Max / Min:     "), data.MaxTemp, data.MinTemp,
		theme.Subtitle.Render("Precip. chance:"), data.PrecProb,
		theme.Subtitle.Render("Cloud cover:   "), data.CloudCover,
	)

	leftAlignedMetrics := lipgloss.NewStyle().Align(lipgloss.Left).Render(rawMetrics)
	metricsBlock := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(leftAlignedMetrics)
	content := lipgloss.JoinVertical(lipgloss.Left, header, "", asciiArt, "", metricsBlock)

	return theme.Card.Width(width).Render(content)
}
