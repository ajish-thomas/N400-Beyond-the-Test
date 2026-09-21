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
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Kind           string   `json:"kind"`
	Source         string   `json:"source"`
	Citation       string   `json:"citation,omitempty"`
	Questions      []int    `json:"questions"`
	SourceStart    int      `json:"source_start"`
	SourceEnd      int      `json:"source_end"`
	EditorialNotes []string `json:"editorial_notes"`
	Blocks         []Block  `json:"-"`
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
	return doc, nil
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
