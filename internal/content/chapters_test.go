package content

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

var updateChapterGolden = flag.Bool("update-chapter-golden", false, "update reviewed chapter fixture")

func TestChapterGolden(t *testing.T) {
	c, err := LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	for _, chapter := range c.Chapters {
		t.Run(chapter.ID, func(t *testing.T) {
			// Blocks are runtime fields, deliberately excluded from front matter JSON.
			fixture := struct {
				Chapter Chapter
				Blocks  []Block
			}{chapter, chapter.Blocks}
			got, err := json.MarshalIndent(fixture, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, '\n')
			file := "testdata/" + chapter.ID + ".golden.json"
			if *updateChapterGolden {
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
				t.Fatal("chapter differs from reviewed golden; inspect original source before updating")
			}
		})
	}
	if len(c.Chapters) != 7 || len(c.Chapters[0].Objectives) != 4 || len(c.Chapters[0].Questions) != 25 || len(c.Chapters[1].Questions) != 20 || len(c.Chapters[2].Questions) != 12 || len(c.Chapters[3].Questions) != 8 || len(c.Chapters[4].Questions) != 14 || len(c.Chapters[5].Questions) != 5 || len(c.Chapters[6].Questions) != 4 || len(c.Images) != 7 {
		t.Fatal("unexpected published chapter inventory")
	}
}

// A word inventory per source page catches omissions and additions even when
// two-column prose must be reordered. The reviewed golden separately pins order.
func TestChapterSourceCoverage(t *testing.T) {
	c, err := LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("data/raw/study-guide.txt")
	if err != nil {
		t.Fatal(err)
	}
	pages := strings.Split(string(raw), "\f")
	b, err := os.ReadFile("data/raw/study-guide-visual-supplement.json")
	if err != nil {
		t.Fatal(err)
	}
	var supplement map[int]string
	if err := json.Unmarshal(b, &supplement); err != nil {
		t.Fatal(err)
	}
	for _, ch := range c.Chapters {
		t.Run(ch.ID, func(t *testing.T) {
			authored := map[int]string{ch.SourceStart: strings.Join(ch.Objectives, " ")}
			for _, b := range ch.Blocks {
				authored[b.Page] += " " + b.Text
				if b.Image != nil {
					authored[b.Page] += " " + b.Image.Credit
				}
				for _, item := range b.Items {
					authored[b.Page] += " " + item.Text + " " + strings.Join(item.Children, " ")
				}
			}
			words := regexp.MustCompile(`[\p{L}\p{N}]+`)
			inventory := func(text string) map[string]int {
				m := map[string]int{}
				for _, word := range words.FindAllString(strings.ToLower(text), -1) {
					m[word]++
				}
				return m
			}
			number := regexp.MustCompile(`^\d+$`)
			// Running chapter-banner headers (e.g. "CHAPTER 4" or "CHAPTER 4: THE
			// JUDICIAL BRANCH" or, per Chapter 6's inconsistent formatting, "CHAPTER
			// 6 – U.S. GEOGRAPHY") are page furniture, not body content. A stray
			// banner from an adjacent chapter can bleed onto a page (confirmed on
			// page 29); this matches any chapter number and any title separator, not
			// just this chapter's own colon-separated form, so those artifacts are
			// excluded too. The chapter title itself is only page furniture on the
			// opening splash page: a later page can legitimately reuse the exact
			// title text as a genuine section heading (chapter 4's page 31), which
			// must still be counted.
			banner := regexp.MustCompile(`^CHAPTER \d+\b.*$`)
			// Chapter 6's maps print several labels as curved or diagonal text
			// (following a mountain range, a river, or a diagonal leader line to a
			// small map marker). pdftotext extracts each as scrambled letter-run
			// fragments, sometimes sharing a line with real prose (e.g. "...in the
			// eastern                tains"). Each affected raw line is mapped here,
			// per page, to its real-prose-only equivalent (dropped entirely if the
			// line is pure fragment); the clean labels themselves are supplied via
			// the visual supplement, same as Chapter 2's page 22 diagram.
			mapLabelFixes := map[int]map[string]string{
				39: {
					// The page 39 map's "Washington, D.C." marker label is split
					// across a diagonal leader line into two fragments.
					"gton, D.C.": "", "Washin": "",
				},
				40: {
					// The page 40 map's "GULF OF AMERICA" label follows a curve.
					"ICA": "", "ME R": "", "GULF OF A": "",
				},
				41: {
					"oun": "", "M": "", "so": "", "yM": "", "is": "", "ur": "", "ver": "",
					"iR": "", "Ri": "", "Ro": "", "siss": "", "Mis i pp i": "", "iver": "",
					"cia": "", "nM": "", "s": "", "ala": "", "pp": "",
					"States. The Appalachian Mountains are in the eastern                                   tains":                                          "States. The Appalachian Mountains are in the eastern",
					"United States, and the Rocky Mountains are in the                               ck":                                                    "United States, and the Rocky Mountains are in the",
					"There are also many rivers in the United States. The                                                                                                         tain": "There are also many rivers in the United States. The",
					"two longest rivers in the U.S. are the Mississippi                                                           A":                                                "two longest rivers in the U.S. are the Mississippi",
				},
				45: {
					// Chapter 7's page 45 "Atlantic Ocean" map label also follows a curve.
					"Plymouth                      At":                     "Plymouth",
					"Jamestown                          l      Portugal":   "Jamestown                          Portugal",
					"ean": "", "Oc": "", "an": "",
					"tic                              Africa": "Africa",
				},
			}
			for page := ch.SourceStart; page <= ch.SourceEnd; page++ {
				var source []string
				for _, line := range strings.Split(pages[page-1], "\n") {
					line = strings.TrimSpace(line)
					if fixes, ok := mapLabelFixes[page]; ok {
						if clean, ok := fixes[line]; ok {
							line = clean
						}
					}
					if number.MatchString(line) || strings.Contains(line, "ONE NATION, ONE PEOPLE: THE USCIS CIVICS TEST TEXTBOOK") || banner.MatchString(line) || (page == ch.SourceStart && line == strings.ToUpper(ch.Title)) || line == "In this chapter, you will learn about:" {
						continue
					}
					source = append(source, line)
				}
				sourceText := strings.Join(source, " ")
				if page == 18 {
					sourceText = strings.ReplaceAll(sourceText, "representa- tives", "representatives")
				}
				if page == 41 {
					// Chapter 6's page 41 independently duplicates several map labels
					// in an ALL-CAPS form layered behind the visible mixed-case
					// labels (verified against the rendered page: only one instance
					// of each is actually printed). One duplicate of each is dropped
					// before comparison rather than transcribed twice.
					for _, dup := range []string{"ALASKA", "HAWAII", "NORTHERN", "MARIANAS", "AMERICAN", "SAMOA", "VIRGIN", "GUAM", "PUERTO", "RICO", "ISLANDS", "ISLANDS"} {
						sourceText = strings.Replace(sourceText, dup, "", 1)
					}
				}
				want, got := inventory(sourceText+" "+supplement[page]), inventory(authored[page])
				var differences []string
				for word, n := range want {
					if got[word] != n {
						differences = append(differences, fmt.Sprintf("%s: source=%d chapter=%d", word, n, got[word]))
					}
				}
				for word, n := range got {
					if want[word] == 0 {
						differences = append(differences, fmt.Sprintf("%s: source=0 chapter=%d", word, n))
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

func TestChapterReferences(t *testing.T) {
	c, err := LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	for _, ch := range c.Chapters {
		for _, id := range ch.Questions {
			if !slices.Contains(c.Questions[id-1].Chapters, ch.ID) {
				t.Errorf("Q%d missing reverse chapter link", id)
			}
		}
	}
	for _, q := range c.Questions {
		for _, id := range q.Chapters {
			found := false
			for _, ch := range c.Chapters {
				if ch.ID == id && slices.Contains(ch.Questions, q.ID) {
					found = true
				}
			}
			if !found {
				t.Errorf("Q%d has an invalid chapter backlink %s", q.ID, id)
			}
		}
	}
	if !slices.Equal(c.Questions[17].Chapters, []string{"constitution", "legislative"}) {
		t.Fatal("Q18 should link to both published chapters")
	}
	if !slices.Equal(c.Questions[40].Chapters, []string{"constitution", "legislative", "executive"}) {
		t.Fatal("Q41 should link to all three published chapters")
	}
	if !slices.Equal(c.Questions[49].Chapters, []string{"constitution", "judicial"}) {
		t.Fatal("Q50 should link to both chapters covering the judicial branch")
	}
	if !slices.Equal(c.Questions[12].Chapters, []string{"constitution", "executive", "rights"}) {
		t.Fatal("Q13 (rule of law) should link to every chapter that repeats it")
	}
	if !slices.Equal(c.Questions[63].Chapters, []string{"legislative", "rights"}) {
		t.Fatal("Q64 should link to both chapters covering federal-office citizenship")
	}
	if !slices.Equal(c.Questions[80].Chapters, []string{"constitution", "geography", "early-history"}) {
		t.Fatal("Q81 should link to all three chapters covering the 13 original states")
	}
	if len(c.Questions[0].Chapters) != 0 {
		t.Fatal("uncovered question given speculative chapter link")
	}
	for _, mutate := range []struct {
		name string
		fn   func(*Catalog)
	}{
		{"unknown question", func(c *Catalog) { c.Chapters[0].Questions[0] = 129 }},
		{"duplicate question", func(c *Catalog) { c.Chapters[0].Questions[1] = c.Chapters[0].Questions[0] }},
		{"unknown image", func(c *Catalog) { c.Chapters[0].Images[0] = "missing" }},
		{"missing credit", func(c *Catalog) { c.Images[0].Credit = "" }},
		{"AP credit", func(c *Catalog) { c.Images[0].Credit = "Courtesy of the Associated Press." }},
		{"unreviewed credit", func(c *Catalog) { c.Images[0].Credit = "AP Photo / Someone" }},
		{"caption drift", func(c *Catalog) { c.Images[0].Caption = "Changed" }},
		{"page drift", func(c *Catalog) { c.Images[0].Page = 8 }},
		{"missing file", func(c *Catalog) { c.Images[0].File = "missing.jpg" }},
		{"unknown chapter", func(c *Catalog) { c.Images[0].Chapters = []string{"missing"} }},
		{"duplicate chapter", func(c *Catalog) { c.Chapters = append(c.Chapters, c.Chapters[0]) }},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			c, err := LoadCatalog()
			if err != nil {
				t.Fatal(err)
			}
			mutate.fn(c)
			if err := c.validate(data); err == nil {
				t.Fatal("accepted invalid catalog")
			}
		})
	}
	files := fstest.MapFS{}
	if err := fs.WalkDir(data, "data/images", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := data.ReadFile(name)
		if err != nil {
			return err
		}
		files[name] = &fstest.MapFile{Data: b}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	files["data/images/orphan.jpg"] = &fstest.MapFile{Data: []byte("unreviewed")}
	if err := c.validate(files); err == nil {
		t.Fatal("accepted orphan image")
	}
	delete(files, "data/images/orphan.jpg")
	files["data/images/ch01-signing.jpg"] = &fstest.MapFile{Data: []byte("corrupt")}
	if err := c.validate(files); err == nil {
		t.Fatal("accepted corrupt image")
	}
}

func TestChapterMalformedInput(t *testing.T) {
	b, err := data.ReadFile("data/chapters/01-constitution.md")
	if err != nil {
		t.Fatal(err)
	}
	source := string(b)
	for _, bad := range []string{
		strings.Replace(source, `"id": "constitution"`, `"unknown": "constitution"`, 1),
		strings.Replace(source, "<!-- page:9 -->", "<!-- page:7 -->", 1),
		strings.Replace(source, "<!-- page:8 -->", "", 1),
		strings.Replace(source, "```text", "```html", 1),
		strings.Replace(source, ":::sidebar", "<script>alert(1)</script>", 1),
		source + "\n```text\nunclosed",
	} {
		if _, err := ParseChapter(strings.NewReader(bad)); err == nil {
			t.Error("accepted malformed chapter")
		}
	}
}
