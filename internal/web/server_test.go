package web

import (
	"bytes"
	"flag"
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
		{"/learn", 200, "2 of 12 chapters"},
		{"/learn/legislative", 200, "How Congress Makes a Federal Law"},
		{"/learn/constitution", 200, "The U.S. Constitution was written in 1787."},
		{"/learn/missing", 404, "404"},
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

var updateReaderGolden = flag.Bool("update-reader-golden", false, "update reviewed reader HTML snapshot")

func TestChapterReaderGoldenAndImages(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"constitution", "legislative"} {
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
	for _, tc := range []struct{ file, kind string }{{"ch01-signing.jpg", "image/jpeg"}, {"ch01-constitution.jpg", "image/jpeg"}, {"ch01-treaty.png", "image/png"}, {"ch02-oval-office.png", "image/png"}} {
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
