package content

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
)

var updateLibraryGolden = flag.Bool("update-library-golden", false, "update reviewed library fixture")

func TestLibraryGolden(t *testing.T) {
	c, err := LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	for _, doc := range c.Library {
		t.Run(doc.ID, func(t *testing.T) {
			fixture := struct {
				Doc    LibraryDoc
				Blocks []Block
			}{doc, doc.Blocks}
			got, err := json.MarshalIndent(fixture, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, '\n')
			file := "testdata/" + doc.ID + ".library.golden.json"
			if *updateLibraryGolden {
				if err := os.MkdirAll("testdata", 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(file, got, 0644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatal("library document differs from reviewed golden; inspect original source before updating")
			}
		})
	}
	if len(c.Library) != 1 || c.Library[0].ID != "declaration" || len(c.Library[0].Questions) != 5 {
		t.Fatal("unexpected published library inventory")
	}
}

// A word inventory per source page catches omissions and additions, the same
// technique TestChapterSourceCoverage uses for the Study Guide.
func TestLibrarySourceCoverage(t *testing.T) {
	c, err := LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	raw := map[string][]string{}
	for _, source := range []string{"constitution", "almanac"} {
		b, err := os.ReadFile("data/raw/" + source + ".txt")
		if err != nil {
			t.Fatal(err)
		}
		raw[source] = strings.Split(string(b), "\f")
	}
	sourceFile := map[string]string{"declaration-constitution": "constitution", "citizens-almanac": "almanac"}
	words := regexp.MustCompile(`[\p{L}\p{N}]+`)
	inventory := func(text string) map[string]int {
		m := map[string]int{}
		for _, word := range words.FindAllString(strings.ToLower(text), -1) {
			m[word]++
		}
		return m
	}
	number := regexp.MustCompile(`^\d+$`)
	for _, doc := range c.Library {
		t.Run(doc.ID, func(t *testing.T) {
			pages := raw[sourceFile[doc.Source]]
			authored := map[int]string{}
			for _, b := range doc.Blocks {
				authored[b.Page] += " " + b.Text
				for _, item := range b.Items {
					authored[b.Page] += " " + item.Text + " " + strings.Join(item.Children, " ")
				}
			}
			for page := doc.SourceStart; page <= doc.SourceEnd; page++ {
				var source []string
				for _, line := range strings.Split(pages[page-1], "\n") {
					line = strings.TrimSpace(line)
					if number.MatchString(line) {
						continue
					}
					// The document's own title, printed as a running head on its
					// opening page, is page furniture already carried by the
					// document's Title field, the same treatment TestChapterSourceCoverage
					// gives a chapter's opening title splash.
					if page == doc.SourceStart && line == strings.ToUpper(doc.Title) {
						continue
					}
					source = append(source, line)
				}
				want, got := inventory(strings.Join(source, " ")), inventory(authored[page])
				var differences []string
				for word, n := range want {
					if got[word] != n {
						differences = append(differences, fmt.Sprintf("%s: source=%d doc=%d", word, n, got[word]))
					}
				}
				for word, n := range got {
					if want[word] == 0 {
						differences = append(differences, fmt.Sprintf("%s: source=0 doc=%d", word, n))
					}
				}
				sort.Strings(differences)
				if len(differences) > 0 {
					t.Errorf("page %d word coverage differs:\n%s", page, strings.Join(differences, "\n"))
				}
			}
		})
	}
}

func TestLibraryReferences(t *testing.T) {
	c, err := LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	for _, doc := range c.Library {
		for _, id := range doc.Questions {
			if !slices.Contains(c.Questions[id-1].Library, doc.ID) {
				t.Errorf("Q%d missing reverse library link", id)
			}
		}
	}
	for _, q := range c.Questions {
		for _, id := range q.Library {
			found := false
			for _, doc := range c.Library {
				if doc.ID == id && slices.Contains(doc.Questions, q.ID) {
					found = true
				}
			}
			if !found {
				t.Errorf("Q%d has an invalid library backlink %s", q.ID, id)
			}
		}
	}
	if !slices.Equal(c.Questions[7].Library, []string{"declaration"}) {
		t.Fatal("Q8 (why the Declaration matters) should link to the Declaration")
	}
	if len(c.Questions[77].Library) != 0 {
		t.Fatal("Q78 (who wrote the Declaration) should not be linked; authorship is not stated on the document's own pages")
	}
}

func TestLibraryMalformedInput(t *testing.T) {
	valid := `---
{"id": "x", "title": "X", "kind": "founding-document", "source": "declaration-constitution", "questions": [], "source_start": 1, "source_end": 1}
---
<!-- page:1 -->
Body text.
`
	for _, tc := range []struct {
		name string
		doc  string
	}{
		{"missing citation", `---
{"id": "x", "title": "X", "kind": "speech", "source": "citizens-almanac", "questions": [], "source_start": 1, "source_end": 1}
---
<!-- page:1 -->
Body text.
`},
		{"unexpected citation", `---
{"id": "x", "title": "X", "kind": "founding-document", "source": "declaration-constitution", "citation": "Some citation.", "questions": [], "source_start": 1, "source_end": 1}
---
<!-- page:1 -->
Body text.
`},
		{"unknown kind", strings.Replace(valid, `"kind": "founding-document"`, `"kind": "essay"`, 1)},
		{"unknown source", strings.Replace(valid, `"source": "declaration-constitution"`, `"source": "wikipedia"`, 1)},
		{"bad id", strings.Replace(valid, `"id": "x"`, `"id": "1x"`, 1)},
		{"empty title", strings.Replace(valid, `"title": "X"`, `"title": ""`, 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseLibraryDoc(strings.NewReader(tc.doc)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
	if _, err := ParseLibraryDoc(strings.NewReader(valid)); err != nil {
		t.Fatalf("expected valid document to parse: %v", err)
	}
}
