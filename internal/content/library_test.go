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
	if len(c.Library) != 5 || c.Library[0].ID != "amendments" || len(c.Library[0].Questions) != 13 || c.Library[1].ID != "declaration" || len(c.Library[1].Questions) != 5 || c.Library[2].ID != "patriotic-anthems" || len(c.Library[2].Questions) != 1 || c.Library[3].ID != "patriotic-symbols" || len(c.Library[3].Questions) != 4 || c.Library[4].ID != "us-constitution" || len(c.Library[4].Questions) != 10 {
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
	// The Almanac's decorative script titles and author attribution lines
	// (e.g. "The Star-Spangled Banner (1814) by Francis Scott Key") render as
	// pure vector art with no extractable PDF text layer at all, the same
	// class of gap Chapter 2's page 22 lawmaking diagram has; their clean
	// transcription lives here, keyed by printed page, the same mechanism
	// study-guide-visual-supplement.json uses for the Study Guide.
	b, err := os.ReadFile("data/raw/almanac-visual-supplement.json")
	if err != nil {
		t.Fatal(err)
	}
	var almanacSupplement map[int]string
	if err := json.Unmarshal(b, &almanacSupplement); err != nil {
		t.Fatal(err)
	}
	sourceFile := map[string]string{"declaration-constitution": "constitution", "citizens-almanac": "almanac"}
	// almanac.txt's raw pages are split by the PDF's own page breaks, which
	// include eight unnumbered/roman-numeral front-matter pages (cover, the
	// "Spirit of '76" plate, title page, ISBN notice, two Table of Contents
	// pages, and two "Message from the Director" pages) before the printed
	// Arabic page 1 begins. Library documents are authored with the printed
	// page numbers a reader sees in the book (matching declaration.md's and
	// us-constitution.md's convention, and what "Source page N" shows in the
	// reader), so this offset is applied only when indexing into the raw
	// almanac pages, never to the authored/displayed page numbers themselves.
	sourceOffset := map[string]int{"declaration-constitution": 0, "citizens-almanac": 8}
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
	// The Almanac's narrow two-column layout hyphenates words across a line
	// wrap far more often than the Constitution's single-column pages, and
	// several of those wraps land on a row that also carries the other
	// column's unrelated text (pdftotext -layout dumps a full row of both
	// columns together), so the two halves of one hyphenated word are not
	// always adjacent once every line is joined. Each affected region is
	// replaced here as one exact block spanning both halves and whatever
	// sits between them, the same targeted, exact-match technique as
	// footnoteRefs above and chapters' mapLabelFixes; this changes only
	// which characters carry the words, never adds or removes a word the
	// source doesn't already have.
	almanacTextFixes := map[int]map[string]string{
		9: {
			"B       eginning early in our": "Beginning early in our",
			"sporting events, spoken expres-         important patriotic anthems and sions have always been an impor-        symbols. tant part of American civic life.": "sporting events, spoken expressions important patriotic anthems and have always been an important symbols. part of American civic life.",
		},
		10: {
			"“    T        he Star-Spangled Banner” is the national anthem": "“The Star-Spangled Banner” is the national anthem",
			"the British of- ficers who agreed to release Dr.":              "the British officers who agreed to release Dr.",
		},
		11: {
			"that he began a poem to com-                       United States. In 1916, President memorate the occasion. He wrote                    Woodrow Wilson ordered that the poem to be sung to the popu-                   the song be played at military lar British song,": "that he began a poem to commemorate United States. In 1916, President the occasion. He wrote Woodrow Wilson ordered that the poem to be sung to the popular the song be played at military British song,",
			"The significance and popular- anthem of the United States. ity of the song spread across the": "The significance and popularity anthem of the United States. of the song spread across the",
		},
		13: {
			"A “        merica the Beautiful” was written in 1893": "“America the Beautiful” was written in 1893",
			"Beautiful” first ap- peared in print in":              "Beautiful” first appeared in print in",
		},
		15: {
			"A        s part of an auction held": "As part of an auction held",
			"freedom and oppor- Inauguration of the Statue of Liberty in 1886, tunity. She saw the new statue": "freedom and opportunity Inauguration of the Statue of Liberty in 1886, She saw the new statue",
		},
		20: {
			"T        he Pledge of Allegiance was first published": "The Pledge of Allegiance was first published",
		},
		21: {
			"will be our coun- of America.                             try’s most powerful resource in": "will be our country's of America. most powerful resource in",
		},
		22: {
			"A       s America fought for": "As America fought for",
			"To emphasize the impor- the United States of America.                 tance of the American flag": "To emphasize the importance the United States of America. of the American flag",
		},
		23: {
			"O         n July 30, 1956, President":               "On July 30, 1956, President",
			"Dwight D. Eisenhower ap- proved a Joint Resolution": "Dwight D. Eisenhower approved a Joint Resolution",
			"“In God We Trust” is also en- can be traced back nearly 200 years                                  graved on the wall above the": "“In God We Trust” is also engraved can be traced back nearly 200 years on the wall above the",
			"and in- cluded on the redesigned two-cent": "and included on the redesigned two-cent",
		},
		24: {
			"O         n July 4, 1776, the":                                   "On July 4, 1776, the",
			"of America. Following the ap-":                                   "of America. Following the appointment",
			"pointment of two additional com-":                                "of two additional committees",
			"“out of many, one” in Latin and mittees, each building upon the": "“out of many, one” in Latin and each building upon the",
		},
		25: {
			"The constel-":                        "The constellation",
			"lation represents the fact that the": "represents the fact that the",
			"The eagle alone sup-":                "The eagle alone supports",
			"ports the shield to signify":         "the shield to signify",
			"The color red signi-":                "The color red signifies",
			"fies valor and bravery":              "valor and bravery",
			"purity and inno-":                    "purity and innocence",
			"cence, and the color blue signifies": ", and the color blue signifies",
		},
		26: {
			"U.S. em- bassies worldwide": "U.S. embassies worldwide",
		},
	}
	for _, doc := range c.Library {
		t.Run(doc.ID, func(t *testing.T) {
			pages := raw[sourceFile[doc.Source]]
			offset := sourceOffset[doc.Source]
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
			for _, l := range strings.Split(pages[doc.SourceStart-1+offset], "\n") {
				t := strings.TrimSpace(l)
				if t == "" {
					if len(span) == 0 {
						continue
					}
					break
				}
				// A footnote-reference digit can be glued directly onto the
				// splash's own first word (e.g. "AMENDMENTS12"); apply the
				// same per-page fix used below so the splash still matches
				// the document's title and is recognized as furniture.
				for old, clean := range footnoteRefs[doc.SourceStart] {
					t = strings.Replace(t, old, clean, 1)
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
				for _, line := range strings.Split(pages[page-1+offset], "\n") {
					line = strings.TrimSpace(line)
					if number.MatchString(line) {
						continue
					}
					// Must run before the splashLines check: amendments.md's
					// splash carries a footnote digit glued onto its own
					// first word ("AMENDMENTS12"), fixed to match
					// splashLines' already-fixed keys, so the fix has to
					// land before the lookup, not after.
					for old, clean := range footnoteRefs[page] {
						line = strings.Replace(line, old, clean, 1)
					}
					if page == doc.SourceStart && splashLines[line] {
						continue
					}
					source = append(source, line)
				}
				sourceText := strings.Join(source, " ")
				if doc.Source == "citizens-almanac" {
					for old, clean := range almanacTextFixes[page] {
						sourceText = strings.ReplaceAll(sourceText, old, clean)
					}
					sourceText += " " + almanacSupplement[page]
				}
				want, got := inventory(sourceText), inventory(authored[page])
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
