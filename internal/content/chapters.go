package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Chapter struct {
	ID             string   `json:"id"`
	Number         int      `json:"number"`
	Title          string   `json:"title"`
	Objectives     []string `json:"objectives"`
	Questions      []int    `json:"questions"`
	Images         []string `json:"images"`
	SourceStart    int      `json:"source_start"`
	SourceEnd      int      `json:"source_end"`
	EditorialNotes []string `json:"editorial_notes"`
	Blocks         []Block  `json:"-"`
}

type ListItem struct {
	Text     string
	Children []string
}

// Blocks contain plain text, never pre-trusted HTML. Templates escape every field.
type Block struct {
	Kind    string
	Text    string
	Page    int
	Anchor  string
	Items   []ListItem
	ImageID string
	Image   *Image
}

var pageMarker = regexp.MustCompile(`^<!-- page:(\d+) -->$`)
var imageMarker = regexp.MustCompile(`^!\[(.+)\]\(([a-z0-9-]+)\)$`)
var contentID = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// ParseChapter supports the small authored Markdown subset documented in
// data/chapters/README.txt. JSON front matter is a dependency-free YAML subset.
func ParseChapter(r io.Reader) (Chapter, error) {
	var chapter Chapter
	b, err := io.ReadAll(r)
	if err != nil {
		return chapter, fmt.Errorf("reading chapter: %w", err)
	}
	parts := strings.SplitN(string(b), "---\n", 3)
	if len(parts) != 3 || parts[0] != "" {
		return chapter, fmt.Errorf("chapter requires delimited metadata")
	}
	decoder := json.NewDecoder(strings.NewReader(parts[1]))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&chapter); err != nil {
		return chapter, fmt.Errorf("decoding chapter metadata: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return chapter, fmt.Errorf("chapter metadata contains trailing data")
	}
	if !contentID.MatchString(chapter.ID) || chapter.Number < 1 || chapter.Title == "" || chapter.SourceStart < 1 || chapter.SourceEnd < chapter.SourceStart || len(chapter.Objectives) == 0 {
		return chapter, fmt.Errorf("invalid chapter metadata")
	}
	lines := strings.Split(parts[2], "\n")
	page := 0
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if match := pageMarker.FindStringSubmatch(line); match != nil {
			n, err := strconv.Atoi(match[1])
			if err != nil || n < chapter.SourceStart || n > chapter.SourceEnd || n <= page {
				return chapter, fmt.Errorf("invalid source page %s", match[1])
			}
			page = n
			chapter.Blocks = append(chapter.Blocks, Block{Kind: "page", Page: page, Anchor: fmt.Sprintf("page-%d", page)})
			continue
		}
		if page == 0 {
			return chapter, fmt.Errorf("chapter content requires a source page")
		}
		block := Block{Page: page}
		switch {
		case line == "```text" || line == ":::sidebar":
			closing := "```"
			block.Kind = "diagram"
			if line == ":::sidebar" {
				closing = ":::"
				block.Kind = "sidebar"
			}
			var text []string
			for i++; i < len(lines) && strings.TrimSpace(lines[i]) != closing; i++ {
				text = append(text, lines[i])
			}
			if i == len(lines) || len(text) == 0 {
				return chapter, fmt.Errorf("unclosed or empty %s on page %d", block.Kind, page)
			}
			block.Text = strings.Join(text, "\n")
		case strings.HasPrefix(line, "### "):
			block.Kind = "subheading"
			block.Text = strings.TrimPrefix(line, "### ")
		case strings.HasPrefix(line, "## "):
			block.Kind = "heading"
			block.Text = strings.TrimPrefix(line, "## ")
		case strings.HasPrefix(line, "!["):
			match := imageMarker.FindStringSubmatch(line)
			if match == nil {
				return chapter, fmt.Errorf("invalid image reference on page %d", page)
			}
			block.Kind = "image"
			block.Text = match[1]
			block.ImageID = match[2]
		case strings.HasPrefix(line, "- "):
			block.Kind = "list"
			for ; i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "- "); i++ {
				item := strings.TrimPrefix(strings.TrimSpace(lines[i]), "- ")
				if strings.HasPrefix(lines[i], "  - ") {
					if len(block.Items) == 0 {
						return chapter, fmt.Errorf("nested list without parent on page %d", page)
					}
					j := len(block.Items) - 1
					block.Items[j].Children = append(block.Items[j].Children, item)
				} else {
					block.Items = append(block.Items, ListItem{Text: item})
				}
			}
			i--
		case strings.HasPrefix(line, "> "):
			block.Kind = "caption"
			block.Text = strings.TrimPrefix(line, "> ")
		case strings.HasPrefix(line, "<"), strings.HasPrefix(line, "#"), strings.HasPrefix(line, "```"), strings.HasPrefix(line, ":::"):
			return chapter, fmt.Errorf("unsupported chapter markup on page %d: %q", page, line)
		default:
			block.Kind = "paragraph"
			block.Text = line
		}
		if block.Kind == "heading" || block.Kind == "subheading" {
			block.Anchor = fmt.Sprintf("section-%d", len(chapter.Blocks))
		}
		chapter.Blocks = append(chapter.Blocks, block)
	}
	if len(chapter.Blocks) == 0 {
		return chapter, fmt.Errorf("chapter body is empty")
	}
	return chapter, nil
}

func loadChapters() ([]Chapter, error) {
	files, err := data.ReadDir("data/chapters")
	if err != nil {
		return nil, err
	}
	var chapters []Chapter
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		b, err := data.ReadFile("data/chapters/" + file.Name())
		if err != nil {
			return nil, err
		}
		chapter, err := ParseChapter(bytes.NewReader(b))
		if err != nil {
			return nil, fmt.Errorf("parsing chapter %s: %w", file.Name(), err)
		}
		chapters = append(chapters, chapter)
	}
	if len(chapters) == 0 {
		return nil, fmt.Errorf("no chapters found")
	}
	return chapters, nil
}
