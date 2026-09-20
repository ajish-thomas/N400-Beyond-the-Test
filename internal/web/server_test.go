package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

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
	for _, path := range []string{"/", "/questions/81", "/static/app.css"} {
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
