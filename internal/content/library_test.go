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
	if len(c.Library) != 3 || c.Library[0].ID != "amendments" || len(c.Library[0].Questions) != 13 || c.Library[1].ID != "declaration" || len(c.Library[1].Questions) != 5 || c.Library[2].ID != "us-constitution" || len(c.Library[2].Questions) != 10 {
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
	// The Constitution's own footnote reference digits (e.g. "...Persons.]1",
	// "...taken.5", "...States11") are inlined into this text edition as a
	// parenthetical at the point of reference (see us-constitution.md's
	// editorial notes) rather than kept as a bare digit; the digit itself is
	// otherwise indistinguishable from a real number token to the word-count
	// regex below (`]1` tokenizes as the word "1"), so each known reference is
	// stripped from the raw source line here, the same targeted, exact-match
	// treatment chapters' mapLabelFixes gives PDF-extraction artifacts.
	footnoteRefs := map[int]map[string]string{
		16: {" other Persons.]1 The": " other Persons.] The", "lature thereof,]2 for": "lature thereof,] for"},
		17: {"such Vacancies.]3": "such Vacancies.]"},
		18: {"ay in December,]4 unle": "ay in December,] unle"},
		23: {"ted to be taken.5": "ted to be taken."},
		25: {"Vice President.]6": "Vice President.]"},
		26: {"all be elected.]7": "all be elected.]"},
		29: {"State;—]8 betw": "State;—] betw", "ns or Subjects.]9": "ns or Subjects.]"},
		30: {"our may be due.]10": "our may be due.]"},
		37: {"HE United States11": "HE United States"},
		39: {"AMENDMENTS12": "AMENDMENTS"},
		41: {"Amendment XI.13": "Amendment XI.", "Amendment XII.14": "Amendment XII."},
		42: {"President—]15": "President—]"},
		43: {"Amendment XIII.16": "Amendment XIII.", "Amendment XIV.17": "Amendment XIV."},
		45: {"Amendment XV.18": "Amendment XV.", "Amendment XVI.19": "Amendment XVI.", "Amendment XVII.20": "Amendment XVII."},
		46: {"Amendment XVIII.21": "Amendment XVIII."},
		47: {"Amendment XIX.22": "Amendment XIX.", "Amendment XX.23": "Amendment XX."},
		48: {"Amendment XXI.24": "Amendment XXI."},
		49: {"Amendment XXII.25": "Amendment XXII."},
		50: {"Amendment XXIII.26": "Amendment XXIII.", "Amendment XXIV.27": "Amendment XXIV."},
		51: {"Amendment XXV.28": "Amendment XXV."},
		52: {"Amendment XXVI.29": "Amendment XXVI."},
		53: {"Amendment XXVII.30": "Amendment XXVII."},
	}
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
			// The document's own title, printed as a running head on its opening
			// page, is page furniture already carried by the document's Title
			// field, the same treatment TestChapterSourceCoverage gives a
			// chapter's opening title splash. The splash can span more than one
			// printed line (e.g. "THE CONSTITUTION" / "OF THE UNITED STATES OF
			// AMERICA"); accumulate leading non-blank lines on the opening page
			// until their concatenation matches the title, the same generalized
			// handling chapters use for a multi-line splash.
			splashLines := map[string]bool{}
			var span []string
			for _, l := range strings.Split(pages[doc.SourceStart-1], "\n") {
				t := strings.TrimSpace(l)
				if t == "" {
					if len(span) == 0 {
						continue
					}
					break
				}
				span = append(span, t)
				if strings.Join(span, " ") == strings.ToUpper(doc.Title) {
					break
				}
			}
			if strings.Join(span, " ") == strings.ToUpper(doc.Title) {
				for _, l := range span {
					splashLines[l] = true
				}
			}
			for page := doc.SourceStart; page <= doc.SourceEnd; page++ {
				var source []string
				for _, line := range strings.Split(pages[page-1], "\n") {
					line = strings.TrimSpace(line)
					if number.MatchString(line) {
						continue
					}
					if page == doc.SourceStart && splashLines[line] {
						continue
					}
					for old, clean := range footnoteRefs[page] {
						line = strings.Replace(line, old, clean, 1)
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
	if !slices.Equal(c.Questions[1].Library, []string{"us-constitution"}) {
		t.Fatal("Q2 (supreme law of the land) should link to the Constitution")
	}
	if !slices.Equal(c.Questions[41].Library, []string{"us-constitution"}) {
		t.Fatal("Q42 (Commander in Chief) should link to the Constitution")
	}
	if len(c.Questions[43].Library) != 0 {
		t.Fatal("Q44 (who vetoes bills) should not be linked; the Constitution's text never uses the word veto")
	}
	if len(c.Questions[54].Library) != 0 {
		t.Fatal("Q55 (Supreme Court justices serve for life) should not be linked; the text says 'good Behaviour', not 'for life'")
	}
	if !slices.Equal(c.Questions[39].Library, []string{"amendments"}) {
		t.Fatal("Q40 (presidential succession) should link to the Twenty-Fifth Amendment")
	}
	if !slices.Equal(c.Questions[98].Library, []string{"amendments"}) {
		t.Fatal("Q99 (the abolition of slavery) should link to the Thirteenth Amendment")
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
