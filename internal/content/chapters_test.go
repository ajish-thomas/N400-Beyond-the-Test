package content

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"regexp"
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
	// Blocks are runtime fields, deliberately excluded from front matter JSON.
	fixture := struct {
		Chapter Chapter
		Blocks  []Block
	}{c.Chapters[0], c.Chapters[0].Blocks}
	got, err := json.MarshalIndent(fixture, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	file := "testdata/constitution.golden.json"
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
	if len(c.Chapters) != 1 || len(c.Chapters[0].Objectives) != 4 || len(c.Chapters[0].Questions) != 25 || len(c.Images) != 3 {
		t.Fatal("unexpected Chapter 1 inventory")
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
	ch := c.Chapters[0]
	authored := map[int]string{8: strings.Join(ch.Objectives, " ")}
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
	for page := ch.SourceStart; page <= ch.SourceEnd; page++ {
		var source []string
		for _, line := range strings.Split(pages[page-1], "\n") {
			line = strings.TrimSpace(line)
			if number.MatchString(line) || strings.Contains(line, "ONE NATION, ONE PEOPLE: THE USCIS CIVICS TEST TEXTBOOK") || line == "CHAPTER 1: THE U.S. CONSTITUTION" || line == "CHAPTER 1" || line == "THE U.S. CONSTITUTION" || line == "In this chapter, you will learn about:" {
				continue
			}
			source = append(source, line)
		}
		want, got := inventory(strings.Join(source, " ")), inventory(authored[page])
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
}

func TestChapterReferences(t *testing.T) {
	c, err := LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	for _, ch := range c.Chapters {
		for _, id := range ch.Questions {
			if !reflect.DeepEqual(c.Questions[id-1].Chapters, []string{ch.ID}) {
				t.Errorf("Q%d missing reverse chapter link", id)
			}
		}
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
