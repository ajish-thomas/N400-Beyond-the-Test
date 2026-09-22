package officials

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGovernorClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("format") != "json" || !strings.Contains(r.URL.Query().Get("query"), "wd:Q99") || !strings.Contains(r.URL.Query().Get("query"), "P6") {
			t.Fatal("missing bounded query")
		}
		if r.Header.Get("User-Agent") == "" || strings.HasPrefix(r.Header.Get("User-Agent"), "Go-http-client") {
			t.Fatal("request must identify itself per Wikidata's user-agent policy")
		}
		w.Header().Set("Content-Type", "application/json")
		// Shaped like the real endpoint's response; see fetch_wikidata_test.go.
		_, _ = w.Write([]byte(`{"head":{"vars":["personLabel"]},"results":{"bindings":[{"personLabel":{"xml:lang":"en","type":"literal","value":"Governor Example"}}]}}`))
	}))
	defer server.Close()
	name, err := (GovernorClient{Endpoint: server.URL, HTTP: server.Client()}).Fetch(context.Background(), "Q99")
	if err != nil || name != "Governor Example" {
		t.Fatalf("name=%q err=%v", name, err)
	}
}

func TestGovernorClientRejectsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "no", http.StatusServiceUnavailable) }))
	defer server.Close()
	if _, err := (GovernorClient{Endpoint: server.URL, HTTP: server.Client()}).Fetch(context.Background(), "Q99"); err == nil {
		t.Fatal("expected failure")
	}
}

func TestGovernorClientRejectsEmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":{"bindings":[]}}`))
	}))
	defer server.Close()
	if _, err := (GovernorClient{Endpoint: server.URL, HTTP: server.Client()}).Fetch(context.Background(), "Q99"); err == nil {
		t.Fatal("expected failure for an empty result set")
	}
}

func TestGovernorClientRejectsInvalidQID(t *testing.T) {
	if _, err := (GovernorClient{Endpoint: "http://example.invalid", HTTP: http.DefaultClient}).Fetch(context.Background(), "'; DROP"); err == nil {
		t.Fatal("expected rejection of a non-QID state identifier")
	}
}

func TestGovernorClientRequiresConfiguration(t *testing.T) {
	if _, err := (GovernorClient{}).Fetch(context.Background(), "Q99"); err == nil {
		t.Fatal("expected an error for an unconfigured client")
	}
}
