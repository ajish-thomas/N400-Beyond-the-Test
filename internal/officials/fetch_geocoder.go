package officials

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// CensusGeocoderEndpoint is the production Census Bureau geocoder used by the
// Settings address-disambiguation action. Tests inject their own httptest
// endpoint instead.
const CensusGeocoderEndpoint = "https://geocoding.geo.census.gov/geocoder/geographies/onelineaddress"

// GeocoderClient resolves a congressional district from a full street
// address via the Census Bureau's keyless geocoding service. It is used only
// when a learner opts into address entry to disambiguate a ZIP that spans
// more than one district; the address itself is sent to the Census Bureau
// but never stored anywhere in this app — only the resulting district
// number is kept.
type GeocoderClient struct {
	Endpoint string
	HTTP     *http.Client
}

// Fetch returns the congressional district number (matching the same
// two-digit format the bundled ZIP-to-district crosswalk uses, e.g. "37" or
// "00" for an at-large seat) for the given address. It requires exactly one
// address match and exactly one congressional-district geography layer in
// the response, refusing to guess when either is ambiguous or missing — the
// caller is expected to ask the learner for a more specific address instead.
func (c GeocoderClient) Fetch(ctx context.Context, address string) (string, error) {
	if c.Endpoint == "" || c.HTTP == nil {
		return "", fmt.Errorf("address lookup client is not configured")
	}
	address = strings.TrimSpace(address)
	if address == "" {
		return "", fmt.Errorf("enter a street address")
	}
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return "", fmt.Errorf("parsing geocoder endpoint: %w", err)
	}
	values := u.Query()
	values.Set("address", address)
	values.Set("benchmark", "Public_AR_Current")
	values.Set("vintage", "Current_Current")
	values.Set("format", "json")
	u.RawQuery = values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("creating geocoder request: %w", err)
	}
	response, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("looking up address: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("looking up address: unexpected HTTP status %d", response.StatusCode)
	}
	var payload struct {
		Result struct {
			AddressMatches []struct {
				Geographies map[string][]struct {
					GEOID string `json:"GEOID"`
				} `json:"geographies"`
			} `json:"addressMatches"`
		} `json:"result"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(&payload); err != nil {
		return "", fmt.Errorf("decoding geocoder response: %w", err)
	}
	matches := payload.Result.AddressMatches
	if len(matches) == 0 {
		return "", fmt.Errorf("address was not found; try a more specific address")
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("address is ambiguous; try a more specific address")
	}
	// The congressional-district layer's name is Congress-numbered (e.g.
	// "119th Congressional Districts", later "120th Congressional
	// Districts") and changes on a schedule this client does not track;
	// matching by substring instead of an exact name keeps it working across
	// that rollover without needing its own update.
	var geoids []string
	for name, entries := range matches[0].Geographies {
		if strings.Contains(name, "Congressional District") {
			for _, entry := range entries {
				geoids = append(geoids, entry.GEOID)
			}
		}
	}
	if len(geoids) == 0 {
		return "", fmt.Errorf("address lookup did not return a congressional district")
	}
	if len(geoids) > 1 {
		return "", fmt.Errorf("address lookup returned more than one congressional district")
	}
	geoid := geoids[0]
	if len(geoid) != 4 {
		return "", fmt.Errorf("address lookup returned an unexpected district identifier %q", geoid)
	}
	return geoid[2:], nil
}
