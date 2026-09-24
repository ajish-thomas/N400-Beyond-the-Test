package content

import (
	"bytes"
	"fmt"
	"io"
	"sort"
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
	Images         []string          `json:"images,omitempty"`
	SourceStart    int               `json:"source_start"`
	SourceEnd      int               `json:"source_end"`
	EditorialNotes []string          `json:"editorial_notes"`
	Categories     []LibraryCategory `json:"categories,omitempty"`
	Blocks         []Block           `json:"-"`
	CategoryNav    []CategoryGroup   `json:"-"`
}

// LibraryCategory is optional, editorial grouping and timeline metadata for
// navigation only (e.g. the Bill of Rights vs. the Reconstruction
// Amendments, and each amendment's own ratification year) — it is never
// checked against the source text the way Blocks are, so it is never
// mistaken for part of the transcribed document itself. Items is one entry
// per consecutive "###" subheading (in document order) belonging to this
// group; the item counts across all categories must sum to exactly the
// number of subheadings in the document. Each item's Year is a fact already
// stated in the document's own transcribed text (e.g. "was ratified
// February 7, 1795"), repeated here only so the reader's table of contents
// can show it without re-parsing prose.
type LibraryCategory struct {
	Name  string         `json:"name"`
	Years string         `json:"years"`
	Items []CategoryItem `json:"items"`
}

type CategoryItem struct {
	Year string `json:"year"`
}

// CategoryGroup is Categories resolved against the document's actual
// subheading blocks, ready for the reader's grouped, chronological table of
// contents.
type CategoryGroup struct {
	Name  string
	Years string
	Items []CategoryEntry
}

// CategoryEntry pairs one subheading with its declared year for the
// timeline-style table of contents.
type CategoryEntry struct {
	Anchor string
	Text   string
	Year   string
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
// categories, in order, pairing each with its declared year. It fails closed
// if the item counts don't exactly cover every subheading, so a category
// list can never silently drift out of sync with the document it describes
// (an amendment added, removed, or reordered without updating its category).
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
		if len(c.Items) < 1 {
			return nil, fmt.Errorf("category %q must cover at least one subheading", c.Name)
		}
		if i+len(c.Items) > len(subheadings) {
			return nil, fmt.Errorf("category %q claims more subheadings than the document has", c.Name)
		}
		entries := make([]CategoryEntry, len(c.Items))
		for j, item := range c.Items {
			if item.Year == "" {
				return nil, fmt.Errorf("category %q item %d is missing a year", c.Name, j)
			}
			entries[j] = CategoryEntry{Anchor: subheadings[i+j].Anchor, Text: subheadings[i+j].Text, Year: item.Year}
		}
		groups = append(groups, CategoryGroup{Name: c.Name, Years: c.Years, Items: entries})
		i += len(c.Items)
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
	sort.SliceStable(docs, func(i, j int) bool {
		return LibraryYear(docs[i].ID) < LibraryYear(docs[j].ID)
	})
	return docs, nil
}

// libraryYear is editorial navigation metadata, not source text.
func LibraryYear(id string) int {
	years := map[string]int{"declaration": 1776, "patriotic-symbols": 1776, "us-constitution": 1787, "amendments": 1791, "washington-farewell-address": 1796, "patriotic-anthems": 1814, "lincoln-first-inaugural-address": 1861, "gettysburg-address": 1863, "four-freedoms": 1941, "kennedy-inaugural-address": 1961, "i-have-a-dream": 1963, "remarks-at-the-brandenburg-gate": 1987}
	return years[id]
}
