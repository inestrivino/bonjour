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
	"time"

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

var weatherBaseURL = "https://api.open-meteo.com/v1/forecast"

// FetchWeatherData takes in a latitude and longitude to perform an API call. It returns a WeatherResult type object and may return an error
func fetchWeatherData(lat, lon float64) (*WeatherResult, error) {
	weatherAPIURL := fmt.Sprintf(
		"%s?latitude=%.4f&longitude=%.4f&current=temperature_2m,precipitation_probability,cloud_cover,is_day&daily=temperature_2m_max,temperature_2m_min&forecast_days=1&timezone=auto",
		weatherBaseURL, lat, lon,
	)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(weatherAPIURL)
	if err != nil {
		return nil, fmt.Errorf("error during http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-valid http response: %s", resp.Status)
	}

	var apiData APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiData); err != nil {
		return nil, fmt.Errorf("error while decoding json: %w", err)
	}

	result := &WeatherResult{
		IsDay:       apiData.Current.IsDay == 1,
		CurrentTemp: apiData.Current.Temperature2m,
		CloudCover:  apiData.Current.CloudCover,
		PrecProb:    apiData.Current.PrecipitationProbability,
	}

	if len(apiData.Daily.Temperature2mMax) > 0 {
		result.MaxTemp = apiData.Daily.Temperature2mMax[0]
	}
	if len(apiData.Daily.Temperature2mMin) > 0 {
		result.MinTemp = apiData.Daily.Temperature2mMin[0]
	}

	return result, nil
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
func RenderWeatherData(lat, lon float64, city string, theme *ui.Theme, miniRender bool) string {
	data, err := fetchWeatherData(lat, lon)
	if err != nil {
		if miniRender {
			return theme.ErrorText.Render(fmt.Sprintf("Weather: %s", err.Error()))
		}
		content := fmt.Sprintf("%s\n\n%s",
			theme.ErrorText.Render("Weather Unavailable"),
			theme.Subtitle.Render(err.Error()),
		)
		return theme.Card.BorderForeground(theme.Muted).Render(content)
	}

	if miniRender {
		// Single-line compact view showing essential temperature and location
		return fmt.Sprintf("%s: %.1f°C (H: %.1f°C / L: %.1f°C)",
			theme.Title.Render(city),
			data.CurrentTemp,
			data.MaxTemp,
			data.MinTemp,
		)
	}

	asciiArt := getASCIIArt(data)
	todayDate := time.Now().Format("02/01/2006")

	// Header with current date
	header := theme.Title.Render(fmt.Sprintf("%s in %s", todayDate, city))

	// ASCII art
	artBlock := theme.Banner.Foreground(theme.Secondary).Render(asciiArt)

	// Weather stats
	metrics := fmt.Sprintf(
		"%s %.1f°C\n%s %.1f°C / %.1f°C\n%s %d%%\n%s %d%%",
		theme.Subtitle.Render("Current:       "), data.CurrentTemp,
		theme.Subtitle.Render("Max / Min:     "), data.MaxTemp, data.MinTemp,
		theme.Subtitle.Render("Precip. chance:"), data.PrecProb,
		theme.Subtitle.Render("Cloud cover:   "), data.CloudCover,
	)

	content := fmt.Sprintf("%s\n\n%s\n\n%s", header, artBlock, metrics)

	return theme.Card.Render(content)
}
