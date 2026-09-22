package web

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"n400/internal/content"
	"n400/internal/officials"
	"n400/internal/quiz"
	"n400/internal/store"
)

type testClock struct{ now time.Time }

func (c testClock) Now() time.Time { return c.now }

func TestEmbeddedAssetsFromEmptyDirectory(t *testing.T) {
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Error(err)
		}
	})
	h, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/", "/questions/81", "/static/app.css", "/learn/constitution", "/images/ch01-signing.jpg"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Errorf("%s: status %d outside repository", path, w.Code)
		}
	}
}

func TestRoutes(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path     string
		status   int
		contains string
	}{
		{"/", 200, "Your civics"},
		{"/questions", 200, "128 questions."},
		{"/questions?deck=6520", 200, "20 questions."},
		{"/questions/1", 200, "Republic"},
		{"/questions/128", 200, "Veterans Day"},
		{"/static/app.css", 200, "prefers-color-scheme"},
		{"/static/theme.js", 200, "n400-theme"},
		{"/learn", 200, "12 of 12 chapters"},
		{"/learn/legislative", 200, "How Congress Makes a Federal Law"},
		{"/learn/executive", 200, "Commander in Chief"},
		{"/learn/judicial", 200, "Statue of Lady Justice"},
		{"/learn/rights", 200, "Federalist Papers"},
		{"/learn/geography", 200, "Rocky Mountains"},
		{"/learn/early-history", 200, "Jamestown"},
		{"/learn/revolution", 200, "Thomas Jefferson"},
		{"/learn/new-government", 200, "Louisiana Territory"},
		{"/learn/civil-war", 200, "Emancipation Proclamation"},
		{"/learn/modern-history", 200, "Pearl Harbor"},
		{"/learn/symbols-holidays", 200, "Statue of Liberty"},
		{"/learn/constitution", 200, "The U.S. Constitution was written in 1787."},
		{"/learn/missing", 404, "404"},
		{"/library", 200, "5 documents"},
		{"/library/amendments", 200, "Amendment XXVII"},
		{"/library/declaration", 200, "Button Gwinnett"},
		{"/library/us-constitution", 200, "Alexander Hamilton"},
		{"/library/patriotic-anthems", 200, "I lift my lamp beside the golden door"},
		{"/library/patriotic-symbols", 200, "and justice for all."},
		{"/library/missing", 404, "404"},
		{"/images/missing.jpg", 404, "404"},
		{"/images/manifest.json", 404, "404"},
		{"/missing", 404, "404"}, {"/questions/0", 404, "404"}, {"/questions/129", 404, "404"}, {"/questions/nope", 404, "404"}, {"/static/missing", 404, "404"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
			if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.contains) {
				t.Fatalf("status %d body %s", w.Code, w.Body.String())
			}
			if tc.status == 200 && !strings.HasPrefix(tc.path, "/static/") && w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
				t.Fatal("missing HTML content type")
			}
		})
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/questions", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST status %d", w.Code)
	}
}

func TestPracticeFlow(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/practice", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Official mock") || !strings.Contains(w.Body.String(), "65/20 study set") {
		t.Fatalf("practice picker: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/practice/start", strings.NewReader("mode=6520"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	location := w.Header().Get("Location")
	if w.Code != http.StatusSeeOther || location == "" {
		t.Fatalf("start practice: %d %q", w.Code, location)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", location, nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Question 1 of 10") || !strings.Contains(w.Body.String(), "6 needed to pass") {
		t.Fatalf("practice question: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", location, strings.NewReader("answer=not-an-answer"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK || (!strings.Contains(w.Body.String(), "I did not recognize that wording") && !strings.Contains(w.Body.String(), "This changing answer needs")) || !strings.Contains(w.Body.String(), "Question 2 of 10") {
		t.Fatalf("practice unrecognized answer: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", location, strings.NewReader("override=right"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Marked right by you.") || !strings.Contains(w.Body.String(), "Question 3 of 10") {
		t.Fatalf("practice override: %d %s", w.Code, w.Body.String())
	}
}

func TestBundledFederalAnswerAppearsWithUSCISVerification(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/38", nil))
	for _, text := range []string{"bundled officials snapshot, 2026-09-21", "Donald J. Trump", "https://www.whitehouse.gov/administration/", "https://www.uscis.gov/citizenship/testupdates"} {
		if !strings.Contains(w.Body.String(), text) {
			t.Errorf("federal answer page missing %q", text)
		}
	}
}

func TestStateSettingResolvesCapital(t *testing.T) {
	progress := store.NewFile(t.TempDir() + "/progress.json")
	h, err := NewWithStore(progress, testClock{now: time.Date(2026, time.September, 21, 9, 0, 0, 0, time.UTC)}, "", false, officials.FederalClient{}, officials.GovernorClient{}, officials.GeocoderClient{})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/settings", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "State or District of Columbia") {
		t.Fatalf("settings page: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/settings/state", strings.NewReader("state=CA"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("save state: %d", w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/62", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Sacramento") || !strings.Contains(w.Body.String(), "bundled state snapshot, 2026-09-21") {
		t.Fatalf("state capital answer: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/23", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Alex Padilla") || !strings.Contains(w.Body.String(), "https://www.senate.gov/senators/index.htm?State=") {
		t.Fatalf("senator answer: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/settings/representative", strings.NewReader("representative=Example+Representative"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("save representative: %d", w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/29", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Example Representative") {
		t.Fatalf("representative answer: %d %s", w.Code, w.Body.String())
	}
}

func TestZIPResolvesRepresentativeFromBundledHouseRoster(t *testing.T) {
	progress := store.NewFile(t.TempDir() + "/progress.json")
	h, err := NewWithStore(progress, testClock{now: time.Now()}, "", false, officials.FederalClient{}, officials.GovernorClient{}, officials.GeocoderClient{})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/settings/state", strings.NewReader("state=CA"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("save state: %d", w.Code)
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/settings/zip", strings.NewReader("zip=90008"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("save zip: %d", w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/29", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Sydney Kamlager-Dove") || !strings.Contains(w.Body.String(), "bundled House roster") {
		t.Fatalf("representative answer: %d %s", w.Code, w.Body.String())
	}
}

func TestSettingsAddressDisambiguatesMultiDistrictZIP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"addressMatches":[{"geographies":{"120th Congressional Districts":[{"GEOID":"0642"}]}}]}}`))
	}))
	defer server.Close()
	progress := store.NewFile(t.TempDir() + "/progress.json")
	h, err := NewWithStore(progress, testClock{now: time.Now()}, "", false, officials.FederalClient{}, officials.GovernorClient{}, officials.GeocoderClient{Endpoint: server.URL, HTTP: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/settings/state", strings.NewReader("state=CA"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/settings/zip", strings.NewReader("zip=90002"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("save multi-district zip: %d", w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/29", nil))
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), "Robert Garcia") {
		t.Fatalf("an unresolved multi-district ZIP must not guess a representative: %s", w.Body.String())
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/settings/address", strings.NewReader("address=123+Main+St%2C+Los+Angeles%2C+CA+90002"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("address lookup: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/29", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Robert Garcia") {
		t.Fatalf("representative after address lookup: %d %s", w.Code, w.Body.String())
	}
}

func TestSettingsAddressLookupFailureIsNonFatal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"addressMatches":[]}}`))
	}))
	defer server.Close()
	progress := store.NewFile(t.TempDir() + "/progress.json")
	h, err := NewWithStore(progress, testClock{now: time.Now()}, "", false, officials.FederalClient{}, officials.GovernorClient{}, officials.GeocoderClient{Endpoint: server.URL, HTTP: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	// The address form only appears once a ZIP has already resolved to more
	// than one district candidate, so set one up first to match the real flow.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/settings/zip", strings.NewReader("zip=90002"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("save multi-district zip: %d", w.Code)
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/settings/address", strings.NewReader("address=nowhere"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Could not look up that address") {
		t.Fatalf("address lookup failure: %d %s", w.Code, w.Body.String())
	}
}

func TestSettingsAddressLookupDisabledOffline(t *testing.T) {
	progress := store.NewFile(t.TempDir() + "/progress.json")
	h, err := NewWithStore(progress, testClock{now: time.Now()}, "", true, officials.FederalClient{}, officials.GovernorClient{}, officials.GeocoderClient{})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/settings/address", strings.NewReader("address=123 Main St"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("address lookup must be rejected offline: %d", w.Code)
	}
}

func wikidataFixtureServer(t *testing.T, governor string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query().Get("query")
		switch {
		case strings.Contains(query, "P1308"):
			_, _ = w.Write([]byte(`{"results":{"bindings":[{"office":{"value":"http://www.wikidata.org/entity/Q11696"},"personLabel":{"value":"Live President"}},{"office":{"value":"http://www.wikidata.org/entity/Q11699"},"personLabel":{"value":"Live VP"}},{"office":{"value":"http://www.wikidata.org/entity/Q912994"},"personLabel":{"value":"Live Speaker"}},{"office":{"value":"http://www.wikidata.org/entity/Q11147"},"personLabel":{"value":"Live Chief Justice"}}]}}`))
		case strings.Contains(query, "P6"):
			_, _ = w.Write([]byte(`{"results":{"bindings":[{"personLabel":{"value":"` + governor + `"}}]}}`))
		default:
			http.Error(w, "unrecognized query", http.StatusBadRequest)
		}
	}))
}

func TestSettingsRefreshUpdatesSidecarAndResets(t *testing.T) {
	server := wikidataFixtureServer(t, "Live California Governor")
	defer server.Close()
	progress := store.NewFile(t.TempDir() + "/progress.json")
	sidecarPath := t.TempDir() + "/officials-live.json"
	client := officials.FederalClient{Endpoint: server.URL, HTTP: server.Client()}
	governor := officials.GovernorClient{Endpoint: server.URL, HTTP: server.Client()}
	h, err := NewWithStore(progress, testClock{now: time.Date(2026, time.September, 22, 9, 0, 0, 0, time.UTC)}, sidecarPath, false, client, governor, officials.GeocoderClient{})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/settings/state", strings.NewReader("state=CA"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("save state: %d", w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/settings/refresh", nil))
	if w.Code != http.StatusSeeOther {
		t.Fatalf("refresh: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/settings", nil))
	if !strings.Contains(w.Body.String(), "last refreshed 2026-09-22") {
		t.Fatalf("settings page missing refreshed date: %s", w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/38", nil))
	if !strings.Contains(w.Body.String(), "Live President") || !strings.Contains(w.Body.String(), "local refreshed data, 2026-09-22") {
		t.Fatalf("question page missing refreshed answer: %s", w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/61", nil))
	if !strings.Contains(w.Body.String(), "Live California Governor") {
		t.Fatalf("question page missing refreshed governor: %s", w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/settings/reset-officials", nil))
	if w.Code != http.StatusSeeOther {
		t.Fatalf("reset: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/38", nil))
	if strings.Contains(w.Body.String(), "Live President") {
		t.Fatalf("reset should fall back to the bundled snapshot: %s", w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/61", nil))
	if strings.Contains(w.Body.String(), "Live California Governor") {
		t.Fatalf("reset should clear the refreshed governor too: %s", w.Body.String())
	}
}

func TestSettingsRefreshFailureIsNonFatal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "down", http.StatusServiceUnavailable) }))
	defer server.Close()
	progress := store.NewFile(t.TempDir() + "/progress.json")
	sidecarPath := t.TempDir() + "/officials-live.json"
	h, err := NewWithStore(progress, testClock{now: time.Now()}, sidecarPath, false, officials.FederalClient{Endpoint: server.URL, HTTP: server.Client()}, officials.GovernorClient{}, officials.GeocoderClient{})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/settings/refresh", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Could not refresh") {
		t.Fatalf("refresh failure: %d %s", w.Code, w.Body.String())
	}
}

func TestSettingsRefreshDisabledOffline(t *testing.T) {
	progress := store.NewFile(t.TempDir() + "/progress.json")
	h, err := NewWithStore(progress, testClock{now: time.Now()}, t.TempDir()+"/officials-live.json", true, officials.FederalClient{}, officials.GovernorClient{}, officials.GeocoderClient{})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/settings", nil))
	if strings.Contains(w.Body.String(), "Refresh current officials <") {
		t.Fatalf("offline settings page should not offer a refresh button: %s", w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/settings/refresh", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("refresh must be rejected offline: %d", w.Code)
	}
}

func TestPracticeSavesAnswerHistory(t *testing.T) {
	progress := store.NewFile(t.TempDir() + "/progress.json")
	h, err := NewWithStore(progress, testClock{now: time.Date(2026, time.September, 21, 9, 0, 0, 0, time.UTC)}, "", false, officials.FederalClient{}, officials.GovernorClient{}, officials.GeocoderClient{})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/practice/start", strings.NewReader("mode=6520"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("start practice: %d", w.Code)
	}
	location := w.Header().Get("Location")
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", location, strings.NewReader("answer=not-an-answer"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("answer practice: %d", w.Code)
	}
	data, err := progress.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Questions) != 1 {
		t.Fatalf("saved question records = %#v", data.Questions)
	}
	for _, record := range data.Questions {
		if record.Incorrect != 1 || record.Correct != 0 {
			t.Fatalf("saved practice record = %#v", record)
		}
	}
}

func TestPracticeFeedbackShowsAnswersAfterTerminalMiss(t *testing.T) {
	templates, err := template.ParseFS(assets, "templates/*.gohtml")
	if err != nil {
		t.Fatal(err)
	}
	question := content.Question{ID: 42, Prompt: "Test prompt", Answers: []content.Answer{{Text: "First answer"}, {Text: "Second answer"}}, RequiredCount: 1}
	view := &practiceView{
		Session:  quiz.Session{Questions: []content.Question{question}, Answered: 1, Complete: true},
		Feedback: &quiz.Result{Message: "I did not recognize that wording."},
		Reviewed: &question,
	}
	var body bytes.Buffer
	if err := templates.ExecuteTemplate(&body, "page", page{Title: "Practice test", Practice: true, PracticeSession: view}); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"Correct answers for Question 42", "First answer", "Second answer"} {
		if !strings.Contains(body.String(), text) {
			t.Errorf("terminal feedback missing %q", text)
		}
	}
}

func TestFlashcardFlowPersistsReview(t *testing.T) {
	progress := store.NewFile(t.TempDir() + "/progress.json")
	clock := testClock{now: time.Date(2026, time.September, 21, 9, 0, 0, 0, time.UTC)}
	h, err := NewWithStore(progress, clock, "", false, officials.FederalClient{}, officials.GovernorClient{}, officials.GeocoderClient{})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/flashcards?deck=6520", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "20 due today") || !strings.Contains(w.Body.String(), "Flashcards / Spaced review") {
		t.Fatalf("flashcard deck: %d %s", w.Code, w.Body.String())
	}
	req := httptest.NewRequest("POST", "/flashcards/1", strings.NewReader("result=right"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/flashcards" {
		t.Fatalf("flashcard review: %d %q", w.Code, w.Header().Get("Location"))
	}
	data, err := progress.Load()
	if err != nil {
		t.Fatal(err)
	}
	if data.Cards[1].Box != 2 || data.Questions[1].Correct != 1 || !data.Questions[1].LastAnswered.Equal(clock.now) {
		t.Fatalf("stored review = %#v %#v", data.Cards[1], data.Questions[1])
	}
	if _, err := NewWithStore(progress, nil, "", false, officials.FederalClient{}, officials.GovernorClient{}, officials.GeocoderClient{}); err == nil {
		t.Fatal("nil clock should be rejected")
	}
}

var updateReaderGolden = flag.Bool("update-reader-golden", false, "update reviewed reader HTML snapshot")

func TestChapterReaderGoldenAndImages(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"constitution", "legislative", "executive", "judicial", "rights", "geography", "early-history", "revolution", "new-government", "civil-war", "modern-history", "symbols-holidays"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/"+id, nil))
		file := "testdata/" + id + ".golden.html"
		if *updateReaderGolden {
			if err := os.MkdirAll("testdata", 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(file, w.Body.Bytes(), 0644); err != nil {
				t.Fatal(err)
			}
		}
		want, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(w.Body.Bytes(), want) {
			t.Fatal("reader differs from reviewed golden")
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/constitution", nil))
	for _, s := range []string{"Courtesy of the Library of Congress.", "Courtesy of the National Archives.", "/questions/81", "5 distinct answers required", "https://www.uscis.gov/citizenship/testupdates", "is a called a governor", "Respresentatives"} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("reader missing %q", s)
		}
	}
	for _, tc := range []struct{ file, kind string }{{"ch01-signing.jpg", "image/jpeg"}, {"ch01-constitution.jpg", "image/jpeg"}, {"ch01-treaty.png", "image/png"}, {"ch02-oval-office.png", "image/png"}, {"ch03-voting.png", "image/png"}, {"ch07-jamestown-street.jpg", "image/jpeg"}, {"ch07-pilgrims.png", "image/png"}, {"ch08-washington-command.jpg", "image/jpeg"}, {"ch08-washington-princeton.jpg", "image/jpeg"}, {"ch08-declaration.jpg", "image/jpeg"}, {"ch08-trumbull.png", "image/png"}, {"ch08-independence-hall.jpg", "image/jpeg"}, {"ch08-yorktown.jpg", "image/jpeg"}, {"ext-mitchell-map-revolution.jpg", "image/jpeg"}, {"ch09-constitution.jpg", "image/jpeg"}, {"ch10-cotton.jpg", "image/jpeg"}, {"ch10-douglass.png", "image/png"}, {"ch10-anthony.png", "image/png"}, {"ch10-emancipation.jpg", "image/jpeg"}, {"ch11-rubber-factories.jpg", "image/jpeg"}, {"ch11-wwi-trench.jpg", "image/jpeg"}, {"ch11-great-war-ends.jpg", "image/jpeg"}, {"ch11-depression-breadline.jpg", "image/jpeg"}, {"ch11-migrant-mother.png", "image/png"}, {"ch11-fdr-declaration.png", "image/png"}, {"ch11-pearl-harbor.png", "image/png"}, {"ch11-eisenhower.png", "image/png"}, {"ch11-moon-landing.jpg", "image/jpeg"}, {"ch11-pentagon-flag.jpg", "image/jpeg"}, {"ch12-statue-of-liberty.jpg", "image/jpeg"}, {"ch12-washington-princeton.jpg", "image/jpeg"}, {"ch12-lincoln.png", "image/png"}} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/images/"+tc.file, nil))
		if w.Code != 200 || w.Header().Get("Content-Type") != tc.kind || w.Body.Len() == 0 {
			t.Errorf("image %s not served correctly", tc.file)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/2", nil))
	if !strings.Contains(w.Body.String(), `href="/learn/constitution"`) {
		t.Fatal("missing chapter backlink")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/legislative", nil))
	for _, s := range []string{"Photo by Cecil Stoughton.", "Courtesy of the John F. Kennedy Presidential Library and Museum.", "Congress votes to override the veto", "The bill does NOT become a law", "20 questions"} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("legislative reader missing %q", s)
		}
	}
	if strings.Count(w.Body.String(), `href="https://www.uscis.gov/citizenship/testupdates"`) != 3 {
		t.Fatal("legislative reader must flag all three changing questions")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/18", nil))
	if !strings.Contains(w.Body.String(), `href="/learn/constitution"`) || !strings.Contains(w.Body.String(), `href="/learn/legislative"`) {
		t.Fatal("Q18 missing one of its chapter backlinks")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/executive", nil))
	for _, s := range []string{"Courtesy of the Library of Congress.", "Cabinet-level positions", "Secretary of War (Defense)", "President Donald J. Trump’s Cabinet, 2025.", "12 questions"} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("executive reader missing %q", s)
		}
	}
	if strings.Count(w.Body.String(), `href="https://www.uscis.gov/citizenship/testupdates"`) != 3 {
		t.Fatal("executive reader must flag all three changing questions")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/41", nil))
	for _, path := range []string{"/learn/constitution", "/learn/legislative", "/learn/executive"} {
		if !strings.Contains(w.Body.String(), `href="`+path+`"`) {
			t.Errorf("Q41 missing chapter backlink to %s", path)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/judicial", nil))
	for _, s := range []string{"Statue of Lady Justice", "Courtroom of the Supreme Court of the United States.", "nine justices on the U.S. Supreme Court", "8 questions"} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("judicial reader missing %q", s)
		}
	}
	if strings.Count(w.Body.String(), `href="https://www.uscis.gov/citizenship/testupdates"`) != 1 {
		t.Fatal("judicial reader must flag its one changing question")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/50", nil))
	for _, path := range []string{"/learn/constitution", "/learn/judicial"} {
		if !strings.Contains(w.Body.String(), `href="`+path+`"`) {
			t.Errorf("Q50 missing chapter backlink to %s", path)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/rights", nil))
	for _, s := range []string{"James Madison, Alexander Hamilton, and John Jay", "Civils Rights Act of 1964", "Courtesy of the Polling Place Photo Project.", "14 questions"} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("rights reader missing %q", s)
		}
	}
	if strings.Contains(w.Body.String(), `href="https://www.uscis.gov/citizenship/testupdates"`) {
		t.Fatal("rights reader should flag no changing questions")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/64", nil))
	for _, path := range []string{"/learn/legislative", "/learn/rights"} {
		if !strings.Contains(w.Body.String(), `href="`+path+`"`) {
			t.Errorf("Q64 missing chapter backlink to %s", path)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/geography", nil))
	for _, s := range []string{"Rocky Mountains", "Gulf of America", "Father of Our Country", "5 questions"} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("geography reader missing %q", s)
		}
	}
	if strings.Count(w.Body.String(), `href="https://www.uscis.gov/citizenship/testupdates"`) != 1 {
		t.Fatal("geography reader must flag its one changing question")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/early-history", nil))
	for _, s := range []string{"Courtesy of the National Park Service.", "Courtesy of the Library of Congress.", "enslaved", "1619", "4 questions"} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("early-history reader missing %q", s)
		}
	}
	if strings.Contains(w.Body.String(), `href="https://www.uscis.gov/citizenship/testupdates"`) {
		t.Fatal("early-history reader should flag no changing questions")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/revolution", nil))
	for _, s := range []string{"Courtesy of the National Archives.", "Courtesy of the U.S. Senate.", "Courtesy of the National Park Service.", "Courtesy of the Library of Congress.", "532915", "life", "liberty", "the pursuit of happiness", "11 questions"} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("revolution reader missing %q", s)
		}
	}
	if strings.Contains(w.Body.String(), `href="https://www.uscis.gov/citizenship/testupdates"`) {
		t.Fatal("revolution reader should flag no changing questions")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/81", nil))
	for _, path := range []string{"/learn/constitution", "/learn/geography", "/learn/early-history", "/learn/revolution"} {
		if !strings.Contains(w.Body.String(), `href="`+path+`"`) {
			t.Errorf("Q81 missing chapter backlink to %s", path)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/new-government", nil))
	for _, s := range []string{"Courtesy of the National Archives.", "Constitutional Convention", "Louisiana Territory", "Trail of Tears", "Mexican-American War", "Inuit", "8 questions"} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("new-government reader missing %q", s)
		}
	}
	if strings.Contains(w.Body.String(), `href="https://www.uscis.gov/citizenship/testupdates"`) {
		t.Fatal("new-government reader should flag no changing questions")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/86", nil))
	for _, path := range []string{"/learn/geography", "/learn/revolution", "/learn/new-government", "/learn/symbols-holidays"} {
		if !strings.Contains(w.Body.String(), `href="`+path+`"`) {
			t.Errorf("Q86 missing chapter backlink to %s", path)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/civil-war", nil))
	for _, s := range []string{"Courtesy of the Library of Congress.", "Courtesy of the National Park Service.", "Frederick Douglass", "Susan B. Anthony", "assassinated", "Juneteenth", "7 questions"} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("civil-war reader missing %q", s)
		}
	}
	if strings.Contains(w.Body.String(), `href="https://www.uscis.gov/citizenship/testupdates"`) {
		t.Fatal("civil-war reader should flag no changing questions")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/modern-history", nil))
	for _, s := range []string{"Courtesy of NASA.", "Courtesy of the White House.", "Courtesy of the National Archives.", "Courtesy of the Library of Congress.", "Pearl Harbor", "communism", "Martin Luther King", "September 11", "15 questions"} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("modern-history reader missing %q", s)
		}
	}
	if strings.Contains(w.Body.String(), `href="https://www.uscis.gov/citizenship/testupdates"`) {
		t.Fatal("modern-history reader should flag no changing questions")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/58", nil))
	for _, path := range []string{"/learn/constitution", "/learn/modern-history"} {
		if !strings.Contains(w.Body.String(), `href="`+path+`"`) {
			t.Errorf("Q58 missing chapter backlink to %s", path)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/learn/symbols-holidays", nil))
	for _, s := range []string{"Courtesy of the Library of Congress.", "Courtesy of the U.S. Senate.", "Statue of Liberty", "Ellis Island", "Angel Island", "Star-Spangled Banner", "Martin Luther King", "Father of Our Country", "U.S. Holidays", "Juneteenth", "Thanksgiving", "11 questions"} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("symbols-holidays reader missing %q", s)
		}
	}
	if strings.Contains(w.Body.String(), `<img src="/images/ch12-iwo-jima`) {
		t.Fatal("symbols-holidays reader must not reproduce the Associated Press-credited Iwo Jima photo")
	}
	if strings.Contains(w.Body.String(), `href="https://www.uscis.gov/citizenship/testupdates"`) {
		t.Fatal("symbols-holidays reader should flag no changing questions")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/91", nil))
	for _, path := range []string{"/learn/new-government", "/learn/symbols-holidays"} {
		if !strings.Contains(w.Body.String(), `href="`+path+`"`) {
			t.Errorf("Q91 missing chapter backlink to %s", path)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/122", nil))
	for _, path := range []string{"/learn/geography", "/learn/symbols-holidays"} {
		if !strings.Contains(w.Body.String(), `href="`+path+`"`) {
			t.Errorf("Q122 missing chapter backlink to %s", path)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/126", nil))
	for _, path := range []string{"/learn/civil-war", "/learn/modern-history", "/learn/symbols-holidays"} {
		if !strings.Contains(w.Body.String(), `href="`+path+`"`) {
			t.Errorf("Q126 missing chapter backlink to %s", path)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/128", nil))
	for _, path := range []string{"/learn/modern-history", "/learn/symbols-holidays"} {
		if !strings.Contains(w.Body.String(), `href="`+path+`"`) {
			t.Errorf("Q128 missing chapter backlink to %s", path)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/125", nil))
	if !strings.Contains(w.Body.String(), `href="/learn/symbols-holidays"`) {
		t.Fatal("Q125 missing chapter backlink to /learn/symbols-holidays")
	}
}

func TestLibraryReaderGolden(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"amendments", "declaration", "us-constitution"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/library/"+id, nil))
		file := "testdata/" + id + ".library.golden.html"
		if *updateReaderGolden {
			if err := os.MkdirAll("testdata", 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(file, w.Body.Bytes(), 0644); err != nil {
				t.Fatal(err)
			}
		}
		want, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(w.Body.Bytes(), want) {
			t.Fatal("library reader differs from reviewed golden")
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/library/amendments", nil))
	for _, s := range []string{
		"AMENDMENTS",
		"Congress shall make no law respecting an establishment of religion",
		"The Senate of the United States shall be composed of two Senators from each State",
		"the Vice President shall become President",
		"eighteen years of age or older",
		"No law, varying the compensation",
		"13 questions",
	} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("amendments reader missing %q", s)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/40", nil))
	if !strings.Contains(w.Body.String(), `href="/library/amendments"`) {
		t.Fatal("Q40 missing library backlink to /library/amendments")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/library/declaration", nil))
	for _, s := range []string{
		"WE hold these Truths to be self-evident",
		"Life, Liberty, and the pursuit of Happiness",
		"Georgia", "Connecticut", "JOHN HANCOCK, President",
		"5 questions",
	} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("declaration reader missing %q", s)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/8", nil))
	if !strings.Contains(w.Body.String(), `href="/library/declaration"`) {
		t.Fatal("Q8 missing library backlink to /library/declaration")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/78", nil))
	if strings.Contains(w.Body.String(), `href="/library/declaration"`) {
		t.Fatal("Q78 should not backlink to the Declaration; authorship is not stated there")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/library/us-constitution", nil))
	for _, s := range []string{
		"We the People of the United States",
		"Article. I.",
		"Article III.",
		"Commander in Chief of the Army and Navy",
		"supreme Law of the Land",
		"(Changed by the Seventeenth Amendment.)",
		"Signers of the Constitution",
		"Alexander Hamilton",
		"10 questions",
	} {
		if !strings.Contains(w.Body.String(), s) {
			t.Errorf("us-constitution reader missing %q", s)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/2", nil))
	if !strings.Contains(w.Body.String(), `href="/library/us-constitution"`) {
		t.Fatal("Q2 missing library backlink to /library/us-constitution")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/questions/44", nil))
	if strings.Contains(w.Body.String(), `href="/library/us-constitution"`) {
		t.Fatal("Q44 should not backlink to the Constitution; the word veto never appears in its text")
	}
}

func TestEveryQuestion(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatal(err)
	}
	changing := map[int]bool{23: true, 29: true, 30: true, 38: true, 39: true, 57: true, 61: true, 62: true}
	counts := map[int]int{10: 2, 48: 2, 65: 3, 67: 2, 69: 2, 81: 5, 126: 3}
	for id := 1; id <= 128; id++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", fmt.Sprintf("/questions/%d", id), nil))
		body := w.Body.String()
		if w.Code != 200 || !strings.Contains(body, "I got this right") {
			t.Errorf("Q%d missing question/self-check", id)
		}
		if changing[id] && !strings.Contains(body, `href="https://www.uscis.gov/citizenship/testupdates"`) {
			t.Errorf("Q%d missing permanent verification link", id)
		}
		if n := counts[id]; n > 1 && !strings.Contains(body, fmt.Sprintf("Name %d distinct answers", n)) {
			t.Errorf("Q%d missing cardinality", id)
		}
	}
}
