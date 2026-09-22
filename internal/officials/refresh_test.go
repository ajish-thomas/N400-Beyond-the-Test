package officials

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// wikidataFixtureServer answers both the federal and the per-state governor
// SPARQL queries this package sends, distinguishing them by query text the
// same way the real endpoint is distinguished by its bound variables.
func wikidataFixtureServer(t *testing.T, governor string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query().Get("query")
		switch {
		case strings.Contains(query, "P1308"):
			_, _ = w.Write([]byte(`{"results":{"bindings":[{"office":{"value":"http://www.wikidata.org/entity/Q11696"},"personLabel":{"value":"Live President"}},{"office":{"value":"http://www.wikidata.org/entity/Q11699"},"personLabel":{"value":"Live VP"}},{"office":{"value":"http://www.wikidata.org/entity/Q912994"},"personLabel":{"value":"Live Speaker"}},{"office":{"value":"http://www.wikidata.org/entity/Q11147"},"personLabel":{"value":"Live Chief Justice"}}]}}`))
		case strings.Contains(query, "P6"):
			if governor == "" {
				http.Error(w, "no governor fixture configured", http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(`{"results":{"bindings":[{"personLabel":{"value":"` + governor + `"}}]}}`))
		default:
			http.Error(w, "unrecognized query", http.StatusBadRequest)
		}
	}))
}

func TestRefreshWritesFederalAndGovernorSidecar(t *testing.T) {
	server := wikidataFixtureServer(t, "Live Governor")
	defer server.Close()
	path := filepath.Join(t.TempDir(), "officials-live.json")
	now := time.Date(2026, time.September, 22, 0, 0, 0, 0, time.UTC)
	client := server.Client()
	federal := FederalClient{Endpoint: server.URL, HTTP: client}
	governor := GovernorClient{Endpoint: server.URL, HTTP: client}
	sidecar, err := Refresh(context.Background(), federal, governor, State{Code: "CA", QID: "Q99"}, path, now)
	if err != nil {
		t.Fatal(err)
	}
	if sidecar.AsOf != "2026-09-22" || sidecar.Federal["president"] != "Live President" || sidecar.Governors["CA"] != "Live Governor" {
		t.Fatalf("refresh result = %+v", sidecar)
	}
	saved, err := LoadSidecar(path)
	if err != nil || saved.Federal["president"] != "Live President" || saved.Governors["CA"] != "Live Governor" {
		t.Fatalf("saved sidecar = %+v, err %v", saved, err)
	}
}

func TestRefreshSkipsGovernorWithoutQID(t *testing.T) {
	server := wikidataFixtureServer(t, "")
	defer server.Close()
	path := filepath.Join(t.TempDir(), "officials-live.json")
	client := server.Client()
	sidecar, err := Refresh(context.Background(), FederalClient{Endpoint: server.URL, HTTP: client}, GovernorClient{Endpoint: server.URL, HTTP: client}, State{Code: "DC"}, path, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(sidecar.Governors) != 0 {
		t.Fatalf("D.C. has no governor and no QID; expected no governor entry, got %+v", sidecar.Governors)
	}
}

func TestRefreshPreservesPreviouslyFetchedGovernorForOtherState(t *testing.T) {
	server := wikidataFixtureServer(t, "New Texas Governor")
	defer server.Close()
	path := filepath.Join(t.TempDir(), "officials-live.json")
	if err := SaveSidecar(path, Sidecar{AsOf: "2026-09-01", Federal: map[string]string{"president": "Old President"}, Governors: map[string]string{"CA": "Old California Governor"}}); err != nil {
		t.Fatal(err)
	}
	client := server.Client()
	sidecar, err := Refresh(context.Background(), FederalClient{Endpoint: server.URL, HTTP: client}, GovernorClient{Endpoint: server.URL, HTTP: client}, State{Code: "TX", QID: "Q1439"}, path, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if sidecar.Governors["CA"] != "Old California Governor" || sidecar.Governors["TX"] != "New Texas Governor" {
		t.Fatalf("expected both states' governors preserved: %+v", sidecar.Governors)
	}
}

func TestRefreshFailureLeavesExistingSidecarIntact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "down", http.StatusServiceUnavailable) }))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "officials-live.json")
	if err := SaveSidecar(path, Sidecar{AsOf: "2026-09-01", Source: "Wikidata", Federal: map[string]string{"president": "Old President"}}); err != nil {
		t.Fatal(err)
	}
	client := server.Client()
	if _, err := Refresh(context.Background(), FederalClient{Endpoint: server.URL, HTTP: client}, GovernorClient{Endpoint: server.URL, HTTP: client}, State{}, path, time.Now()); err == nil {
		t.Fatal("expected refresh failure")
	}
	saved, err := LoadSidecar(path)
	if err != nil || saved.AsOf != "2026-09-01" || saved.Federal["president"] != "Old President" {
		t.Fatalf("previous sidecar must survive a failed refresh: %+v, err %v", saved, err)
	}
}

func TestRefreshFailureFromGovernorLeavesExistingSidecarIntact(t *testing.T) {
	server := wikidataFixtureServer(t, "")
	defer server.Close()
	path := filepath.Join(t.TempDir(), "officials-live.json")
	if err := SaveSidecar(path, Sidecar{AsOf: "2026-09-01", Federal: map[string]string{"president": "Old President"}}); err != nil {
		t.Fatal(err)
	}
	client := server.Client()
	if _, err := Refresh(context.Background(), FederalClient{Endpoint: server.URL, HTTP: client}, GovernorClient{Endpoint: server.URL, HTTP: client}, State{Code: "CA", QID: "Q99"}, path, time.Now()); err == nil {
		t.Fatal("expected refresh failure when the governor fetch fails")
	}
	saved, err := LoadSidecar(path)
	if err != nil || saved.Federal["president"] != "Old President" {
		t.Fatalf("previous sidecar must survive a failed governor refresh: %+v, err %v", saved, err)
	}
}
