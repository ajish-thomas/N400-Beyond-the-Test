package officials

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFederalClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("format") != "json" || !strings.Contains(r.URL.Query().Get("query"), "Q11696") {
			t.Fatal("missing bounded query")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":{"bindings":[{"office":{"value":"http://www.wikidata.org/entity/Q11696"},"personLabel":{"value":"President"}},{"office":{"value":"http://www.wikidata.org/entity/Q11699"},"personLabel":{"value":"Vice President"}},{"office":{"value":"http://www.wikidata.org/entity/Q912994"},"personLabel":{"value":"Speaker"}},{"office":{"value":"http://www.wikidata.org/entity/Q11147"},"personLabel":{"value":"Chief Justice"}}]}}`))
	}))
	defer server.Close()
	answers, err := (FederalClient{Endpoint: server.URL, HTTP: server.Client()}).Fetch(context.Background())
	if err != nil || answers["president"] != "President" || len(answers) != 4 {
		t.Fatalf("answers=%v err=%v", answers, err)
	}
}

func TestFederalClientRejectsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "no", http.StatusServiceUnavailable) }))
	defer server.Close()
	if _, err := (FederalClient{Endpoint: server.URL, HTTP: server.Client()}).Fetch(context.Background()); err == nil {
		t.Fatal("expected failure")
	}
}
