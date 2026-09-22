package officials

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func federalFixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":{"bindings":[{"office":{"value":"http://www.wikidata.org/entity/Q11696"},"personLabel":{"value":"Live President"}},{"office":{"value":"http://www.wikidata.org/entity/Q11699"},"personLabel":{"value":"Live VP"}},{"office":{"value":"http://www.wikidata.org/entity/Q912994"},"personLabel":{"value":"Live Speaker"}},{"office":{"value":"http://www.wikidata.org/entity/Q11147"},"personLabel":{"value":"Live Chief Justice"}}]}}`))
	}))
}

func TestRefreshWritesSidecar(t *testing.T) {
	server := federalFixtureServer(t)
	defer server.Close()
	path := filepath.Join(t.TempDir(), "officials-live.json")
	now := time.Date(2026, time.September, 22, 0, 0, 0, 0, time.UTC)
	sidecar, err := Refresh(context.Background(), FederalClient{Endpoint: server.URL, HTTP: server.Client()}, path, now)
	if err != nil {
		t.Fatal(err)
	}
	if sidecar.AsOf != "2026-09-22" || sidecar.Federal["president"] != "Live President" {
		t.Fatalf("refresh result = %+v", sidecar)
	}
	saved, err := LoadSidecar(path)
	if err != nil || saved.AsOf != "2026-09-22" || saved.Federal["president"] != "Live President" {
		t.Fatalf("saved sidecar = %+v, err %v", saved, err)
	}
}

func TestRefreshFailureLeavesExistingSidecarIntact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "down", http.StatusServiceUnavailable) }))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "officials-live.json")
	if err := SaveSidecar(path, Sidecar{AsOf: "2026-09-01", Source: "Wikidata", Federal: map[string]string{"president": "Old President"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Refresh(context.Background(), FederalClient{Endpoint: server.URL, HTTP: server.Client()}, path, time.Now()); err == nil {
		t.Fatal("expected refresh failure")
	}
	saved, err := LoadSidecar(path)
	if err != nil || saved.AsOf != "2026-09-01" || saved.Federal["president"] != "Old President" {
		t.Fatalf("previous sidecar must survive a failed refresh: %+v, err %v", saved, err)
	}
}
