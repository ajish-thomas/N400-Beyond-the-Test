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

// WikidataEndpoint is the production SPARQL endpoint used by the Settings
// refresh action. Tests inject their own httptest endpoint instead.
const WikidataEndpoint = "https://query.wikidata.org/sparql"

// wikidataUserAgent identifies this app to Wikidata's query service, per its
// user-agent policy (https://meta.wikimedia.org/wiki/User-Agent_policy).
// Requests without an identifying header are rejected outright with HTTP 403.
const wikidataUserAgent = "n400-civics-study-app/1.0 (offline-first USCIS civics study tool; local use only)"

// FederalClient fetches only the four changing federal offices. Callers invoke
// it from an explicit refresh action; it is never used during application
// startup or a request that renders a study page.
type FederalClient struct {
	Endpoint string
	HTTP     *http.Client
}

var federalOfficeKeys = map[string]string{
	"Q11696":  "president",
	"Q11699":  "vice_president",
	"Q912994": "speaker",
	"Q11147":  "chief_justice",
}

func (c FederalClient) Fetch(ctx context.Context) (map[string]string, error) {
	if c.Endpoint == "" || c.HTTP == nil {
		return nil, fmt.Errorf("federal refresh client is not configured")
	}
	query := `SELECT ?office ?personLabel WHERE { VALUES ?office { wd:Q11696 wd:Q11699 wd:Q912994 wd:Q11147 } ?office wdt:P1308 ?person . SERVICE wikibase:label { bd:serviceParam wikibase:language "en". } }`
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("parsing refresh endpoint: %w", err)
	}
	values := u.Query()
	values.Set("query", query)
	values.Set("format", "json")
	u.RawQuery = values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating refresh request: %w", err)
	}
	req.Header.Set("Accept", "application/sparql-results+json")
	req.Header.Set("User-Agent", wikidataUserAgent)
	response, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching federal officials: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching federal officials: unexpected HTTP status %d", response.StatusCode)
	}
	// The real SPARQL JSON response nests other standard fields around and
	// inside each binding (a top-level "head", and "type"/"xml:lang" beside
	// "value") that carry no meaning for us and vary by value type; only the
	// fields declared below are read, and DisallowUnknownFields is
	// deliberately not used here for that reason. Correctness is instead
	// enforced below by requiring all four offices to resolve to a value.
	var payload struct {
		Results struct {
			Bindings []struct {
				Office struct {
					Value string `json:"value"`
				} `json:"office"`
				PersonLabel struct {
					Value string `json:"value"`
				} `json:"personLabel"`
			} `json:"bindings"`
		} `json:"results"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decoding federal refresh: %w", err)
	}
	answers := make(map[string]string)
	for _, binding := range payload.Results.Bindings {
		parts := strings.Split(strings.TrimSuffix(binding.Office.Value, "/"), "/")
		key := federalOfficeKeys[parts[len(parts)-1]]
		if key != "" && strings.TrimSpace(binding.PersonLabel.Value) != "" {
			answers[key] = strings.TrimSpace(binding.PersonLabel.Value)
		}
	}
	if len(answers) != len(federalOfficeKeys) {
		return nil, fmt.Errorf("federal refresh returned %d of %d offices", len(answers), len(federalOfficeKeys))
	}
	return answers, nil
}
