package content

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// The Citizen's Almanac's own required citation for any reuse of its text,
// verbatim from AGENTS.md's "Licensing and attribution" section.
const almanacCitation = "U.S. Department of Homeland Security, U.S. Citizenship and Immigration Services, Office of Citizenship, The Citizen's Almanac, Washington, DC, 2014."

var libraryKinds = map[string]bool{"founding-document": true, "speech": true, "symbol": true, "case": true}
var librarySources = map[string]bool{"declaration-constitution": true, "citizens-almanac": true}

type LibraryDoc struct {
	ID             string            `json:"id"`
	Title          string            `json:"title"`
	Kind           string            `json:"kind"`
	Source         string            `json:"source"`
	Citation       string            `json:"citation,omitempty"`
	Questions      []int             `json:"questions"`
	SourceStart    int               `json:"source_start"`
	SourceEnd      int               `json:"source_end"`
	EditorialNotes []string          `json:"editorial_notes"`
	Categories     []LibraryCategory `json:"categories,omitempty"`
	Blocks         []Block           `json:"-"`
	CategoryNav    []CategoryGroup   `json:"-"`
}

// LibraryCategory is optional, editorial grouping metadata for navigation
// only (e.g. the Bill of Rights vs. the Reconstruction Amendments) — it is
// never checked against the source text the way Blocks are, so a category
// name or year range is never mistaken for part of the transcribed document
// itself. Count is how many consecutive "###" subheadings (in document
// order) belong to this group; the counts across all categories must sum to
// exactly the number of subheadings in the document.
type LibraryCategory struct {
	Name  string `json:"name"`
	Years string `json:"years"`
	Count int    `json:"count"`
}

// CategoryGroup is Categories resolved against the document's actual
// subheading blocks, ready for the reader's grouped table of contents.
type CategoryGroup struct {
	Name  string
	Years string
	Items []Block
}

// ParseLibraryDoc supports the same authored Markdown subset as ParseChapter
// (data/chapters/README.txt), for reference-library documents instead of
// Study Guide chapters. See data/library/README.txt.
func ParseLibraryDoc(r io.Reader) (LibraryDoc, error) {
	var doc LibraryDoc
	body, err := splitDocument(r, &doc)
	if err != nil {
		return doc, err
	}
	if !contentID.MatchString(doc.ID) || doc.Title == "" || !libraryKinds[doc.Kind] || !librarySources[doc.Source] || doc.SourceStart < 1 || doc.SourceEnd < doc.SourceStart {
		return doc, fmt.Errorf("invalid library document metadata")
	}
	switch doc.Source {
	case "citizens-almanac":
		if doc.Citation != almanacCitation {
			return doc, fmt.Errorf("document %s: almanac-sourced content requires the exact required citation", doc.ID)
		}
	default:
		if doc.Citation != "" {
			return doc, fmt.Errorf("document %s: citation only applies to almanac-sourced content", doc.ID)
		}
	}
	doc.Blocks, err = parseBlocks(body, doc.SourceStart, doc.SourceEnd)
	if err != nil {
		return doc, err
	}
	if len(doc.Categories) > 0 {
		doc.CategoryNav, err = resolveCategories(doc.Categories, doc.Blocks)
		if err != nil {
			return doc, fmt.Errorf("document %s: %w", doc.ID, err)
		}
	}
	return doc, nil
}

// resolveCategories groups the document's subheadings into its declared
// categories, in order. It fails closed if the counts don't exactly cover
// every subheading, so a category list can never silently drift out of sync
// with the document it describes (an amendment added, removed, or
// reordered without updating its category's count).
func resolveCategories(categories []LibraryCategory, blocks []Block) ([]CategoryGroup, error) {
	var subheadings []Block
	for _, b := range blocks {
		if b.Kind == "subheading" {
			subheadings = append(subheadings, b)
		}
	}
	groups := make([]CategoryGroup, 0, len(categories))
	i := 0
	for _, c := range categories {
		if c.Count < 1 {
			return nil, fmt.Errorf("category %q must cover at least one subheading", c.Name)
		}
		if i+c.Count > len(subheadings) {
			return nil, fmt.Errorf("category %q claims more subheadings than the document has", c.Name)
		}
		groups = append(groups, CategoryGroup{Name: c.Name, Years: c.Years, Items: subheadings[i : i+c.Count]})
		i += c.Count
	}
	if i != len(subheadings) {
		return nil, fmt.Errorf("categories cover %d of %d subheadings", i, len(subheadings))
	}
	return groups, nil
}

func loadLibrary() ([]LibraryDoc, error) {
	files, err := data.ReadDir("data/library")
	if err != nil {
		return nil, err
	}
	var docs []LibraryDoc
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		b, err := data.ReadFile("data/library/" + file.Name())
		if err != nil {
			return nil, err
		}
		doc, err := ParseLibraryDoc(bytes.NewReader(b))
		if err != nil {
			return nil, fmt.Errorf("parsing library document %s: %w", file.Name(), err)
		}
		docs = append(docs, doc)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no library documents found")
	}
	return docs, nil
}
