package officials

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var wikidataQID = regexp.MustCompile(`^Q[0-9]+$`)

// GovernorClient fetches the current governor of one state, identified by its
// Wikidata QID (P6, head of government, on the state's own entity). Callers
// invoke it only from an explicit refresh action, for whichever single state
// the learner has selected — never for all fifty at once.
type GovernorClient struct {
	Endpoint string
	HTTP     *http.Client
}

func (c GovernorClient) Fetch(ctx context.Context, stateQID string) (string, error) {
	if c.Endpoint == "" || c.HTTP == nil {
		return "", fmt.Errorf("governor refresh client is not configured")
	}
	if !wikidataQID.MatchString(stateQID) {
		return "", fmt.Errorf("invalid state identifier %q", stateQID)
	}
	query := fmt.Sprintf(`SELECT ?personLabel WHERE { wd:%s wdt:P6 ?person . SERVICE wikibase:label { bd:serviceParam wikibase:language "en". } }`, stateQID)
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return "", fmt.Errorf("parsing refresh endpoint: %w", err)
	}
	values := u.Query()
	values.Set("query", query)
	values.Set("format", "json")
	u.RawQuery = values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("creating refresh request: %w", err)
	}
	req.Header.Set("Accept", "application/sparql-results+json")
	req.Header.Set("User-Agent", wikidataUserAgent)
	response, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching governor: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching governor: unexpected HTTP status %d", response.StatusCode)
	}
	// See FederalClient.Fetch for why DisallowUnknownFields is deliberately
	// not used here: the real response nests other standard SPARQL fields
	// that carry no meaning for us. Correctness is enforced below by
	// requiring a non-empty result.
	var payload struct {
		Results struct {
			Bindings []struct {
				PersonLabel struct {
					Value string `json:"value"`
				} `json:"personLabel"`
			} `json:"bindings"`
		} `json:"results"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(&payload); err != nil {
		return "", fmt.Errorf("decoding governor refresh: %w", err)
	}
	if len(payload.Results.Bindings) == 0 {
		return "", fmt.Errorf("governor refresh returned no result for %s", stateQID)
	}
	name := strings.TrimSpace(payload.Results.Bindings[0].PersonLabel.Value)
	if name == "" {
		return "", fmt.Errorf("governor refresh returned an empty name for %s", stateQID)
	}
	return name, nil
}
