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
