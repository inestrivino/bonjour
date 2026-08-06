package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/text/language"
)

// LocationResult is all the information about the place that is being geocoded
type LocationResult struct {
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Admin1      string  `json:"admin1"`
}

// geocodingResponse is the list of results given by the API
type geocodingResponse struct {
	Results []LocationResult `json:"results"`
}

// ResolveLocation receives the city name, countryInput, admin1 and returns a LocationResult object
func ResolveLocation(city, countryInput, admin1 string) (*LocationResult, error) {
	city = strings.TrimSpace(city)
	if city == "" {
		return nil, nil
	}

	//parameters for the API call, we want to send the city name, receive max 10 results, in english, and in json format
	params := url.Values{}
	params.Set("name", city)
	params.Set("count", "10")
	params.Set("language", "en")
	params.Set("format", "json")

	//Convert the country name into a country code for the API call
	countryCode := NormalizeCountryCode(countryInput)
	if countryCode != "" {
		params.Set("countryCode", countryCode)
	}

	//the endpoint is defined based on the params
	endpoint := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?%s", params.Encode())

	//we set a 10 second timeout
	client := &http.Client{Timeout: 10 * time.Second}

	//We create a new GET request with a custom header User-Agent so our call is not de-prioratized
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "bonjour/1.0")

	// We execute the get request and analyze the response for errors
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocoding lookup timed out. Check your internet connection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding service returned status %d", resp.StatusCode)
	}

	// We decode the data from the response
	var data geocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse location response")
	}

	//inform the user if no matching locations were found
	if len(data.Results) == 0 {
		return nil, fmt.Errorf("no matching location found for %q", city)
	}

	//if there were matching results we check that, as long as the administration isn't empty, it matches the result
	trimmedAdmin := strings.TrimSpace(admin1)
	if trimmedAdmin != "" {
		for _, res := range data.Results {
			if strings.EqualFold(res.Admin1, trimmedAdmin) {
				return &res, nil
			}
		}
		return nil, fmt.Errorf("found city %q, but failed to match region/state %q", city, trimmedAdmin)
	}

	//we take the top result
	return &data.Results[0], nil
}

// NormalizeCountryCode converts full country names (e.g., "Spain", "United Kingdom")
// or 2-letter codes (e.g., "es", "ES") into standard 2-letter ISO country codes.
func NormalizeCountryCode(input string) string {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return ""
	}

	// If we receive a 2 letter code, then we return it in upper case
	if len(input) == 2 {
		return strings.ToUpper(input)
	}

	// We use Go's language tag parser to try to find the country code from the name
	tag, err := language.ParseRegion(input)
	if err == nil {
		return tag.String()
	}

	// If the tag came back empty then we try to parse common names to their corresponding codes
	commonMap := map[string]string{
		"spain":          "ES",
		"españa":         "ES",
		"united kingdom": "GB",
		"uk":             "GB",
		"great britain":  "GB",
		"united states":  "US",
		"usa":            "US",
		"france":         "FR",
		"germany":        "DE",
		"deutschland":    "DE",
		"italy":          "IT",
		"italia":         "IT",
	}

	if code, ok := commonMap[input]; ok {
		return code
	}

	return ""
}
