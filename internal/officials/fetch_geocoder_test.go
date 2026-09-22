package officials

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// geocoderFixture mirrors the real Census response's shape: several
// unrelated geography layers alongside the one this client actually reads,
// so a regression that narrows the decode too far fails here too.
const geocoderFixture = `{"result":{"addressMatches":[{"geographies":{"States":[{"GEOID":"24"}],"Counties":[{"GEOID":"24033"}],"120th Congressional Districts":[{"GEOID":"2404","CD120":"04","NAME":"Congressional District 4"}]}}]}}`

func TestGeocoderClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("address") == "" || r.URL.Query().Get("benchmark") == "" || r.URL.Query().Get("vintage") == "" {
			t.Fatal("missing required geocoder parameters")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(geocoderFixture))
	}))
	defer server.Close()
	district, err := (GeocoderClient{Endpoint: server.URL, HTTP: server.Client()}).Fetch(context.Background(), "4600 Silver Hill Rd, Washington, DC 20233")
	if err != nil || district != "04" {
		t.Fatalf("district=%q err=%v", district, err)
	}
}

func TestGeocoderClientRejectsNoMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"addressMatches":[]}}`))
	}))
	defer server.Close()
	if _, err := (GeocoderClient{Endpoint: server.URL, HTTP: server.Client()}).Fetch(context.Background(), "nowhere"); err == nil {
		t.Fatal("expected a not-found error")
	}
}

func TestGeocoderClientRejectsAmbiguousMatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"addressMatches":[{"geographies":{}},{"geographies":{}}]}}`))
	}))
	defer server.Close()
	if _, err := (GeocoderClient{Endpoint: server.URL, HTTP: server.Client()}).Fetch(context.Background(), "123 Main St"); err == nil {
		t.Fatal("expected an ambiguous-match error")
	}
}

func TestGeocoderClientRejectsMissingDistrictLayer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"addressMatches":[{"geographies":{"States":[{"GEOID":"24"}]}}]}}`))
	}))
	defer server.Close()
	if _, err := (GeocoderClient{Endpoint: server.URL, HTTP: server.Client()}).Fetch(context.Background(), "123 Main St"); err == nil || !strings.Contains(err.Error(), "did not return a congressional district") {
		t.Fatalf("expected a missing-layer error, got %v", err)
	}
}

func TestGeocoderClientRejectsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "no", http.StatusServiceUnavailable) }))
	defer server.Close()
	if _, err := (GeocoderClient{Endpoint: server.URL, HTTP: server.Client()}).Fetch(context.Background(), "123 Main St"); err == nil {
		t.Fatal("expected failure")
	}
}

func TestGeocoderClientRejectsEmptyAddress(t *testing.T) {
	if _, err := (GeocoderClient{Endpoint: "http://example.invalid", HTTP: http.DefaultClient}).Fetch(context.Background(), "   "); err == nil {
		t.Fatal("expected an error for a blank address")
	}
}

func TestGeocoderClientRequiresConfiguration(t *testing.T) {
	if _, err := (GeocoderClient{}).Fetch(context.Background(), "123 Main St"); err == nil {
		t.Fatal("expected an error for an unconfigured client")
	}
}
