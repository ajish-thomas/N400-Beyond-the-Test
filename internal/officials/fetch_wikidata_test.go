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
		if r.Header.Get("User-Agent") == "" || strings.HasPrefix(r.Header.Get("User-Agent"), "Go-http-client") {
			t.Fatal("request must identify itself per Wikidata's user-agent policy")
		}
		w.Header().Set("Content-Type", "application/json")
		// Shaped like the real endpoint's response, including the top-level
		// "head" object and the "type"/"xml:lang" fields the real endpoint adds
		// beside every "value", so a regression that breaks real decoding
		// (by re-adding DisallowUnknownFields, say) fails here too.
		_, _ = w.Write([]byte(`{"head":{"vars":["office","personLabel"]},"results":{"bindings":[{"office":{"type":"uri","value":"http://www.wikidata.org/entity/Q11696"},"personLabel":{"xml:lang":"en","type":"literal","value":"President"}},{"office":{"type":"uri","value":"http://www.wikidata.org/entity/Q11699"},"personLabel":{"xml:lang":"en","type":"literal","value":"Vice President"}},{"office":{"type":"uri","value":"http://www.wikidata.org/entity/Q912994"},"personLabel":{"xml:lang":"en","type":"literal","value":"Speaker"}},{"office":{"type":"uri","value":"http://www.wikidata.org/entity/Q11147"},"personLabel":{"xml:lang":"en","type":"literal","value":"Chief Justice"}}]}}`))
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
