package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"path"
	"slices"
	"strings"
)

type Image struct {
	ID       string   `json:"id"`
	File     string   `json:"file"`
	Page     int      `json:"page"`
	Caption  string   `json:"caption"`
	Credit   string   `json:"credit"`
	Chapters []string `json:"chapters"`
	Library  []string `json:"library,omitempty"`
	Width    int      `json:"-"`
	Height   int      `json:"-"`
}

type Catalog struct {
	Questions []Question
	Chapters  []Chapter
	Library   []LibraryDoc
	Images    []Image
}

// LoadCatalog builds both directions of chapter links and validates every
// published reference. Full 128-question chapter coverage awaits all 12 chapters.
func LoadCatalog() (*Catalog, error) {
	questions, err := Load()
	if err != nil {
		return nil, err
	}
	chapters, err := loadChapters()
	if err != nil {
		return nil, err
	}
	library, err := loadLibrary()
	if err != nil {
		return nil, err
	}
	b, err := data.ReadFile("data/images/manifest.json")
	if err != nil {
		return nil, err
	}
	var images []Image
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&images); err != nil {
		return nil, fmt.Errorf("decoding image manifest: %w", err)
	}
	catalog := &Catalog{Questions: questions, Chapters: chapters, Library: library, Images: images}
	if err := catalog.validate(data); err != nil {
		return nil, err
	}
	return catalog, nil
}

func (c *Catalog) validate(files fs.FS) error {
	chapterIDs := map[string]bool{}
	numbers := map[int]bool{}
	for _, ch := range c.Chapters {
		if chapterIDs[ch.ID] || numbers[ch.Number] {
			return fmt.Errorf("duplicate chapter %s", ch.ID)
		}
		chapterIDs[ch.ID] = true
		numbers[ch.Number] = true
	}
	imageIDs := map[string]*Image{}
	imageFiles := map[string]bool{}
	for i := range c.Images {
		img := &c.Images[i]
		if !contentID.MatchString(img.ID) || imageIDs[img.ID] != nil || imageFiles[img.File] {
			return fmt.Errorf("invalid or duplicate image %s", img.ID)
		}
		if img.File != path.Base(img.File) || (path.Ext(img.File) != ".jpg" && path.Ext(img.File) != ".png") {
			return fmt.Errorf("invalid image filename %q", img.File)
		}
		if img.Caption == "" || img.Credit == "" || (len(img.Chapters) == 0 && len(img.Library) == 0) {
			return fmt.Errorf("image %s missing caption, credit, or owner", img.ID)
		}
		credit := strings.ToLower(img.Credit)
		if strings.Contains(credit, "associated press") || strings.EqualFold(strings.TrimSpace(img.Credit), "AP") {
			return fmt.Errorf("image %s has prohibited AP credit", img.ID)
		}
		// Currently only the reviewed federal institutions are allowed. Expand
		// this allowlist only with a source and licensing review.
		if img.Credit != "Courtesy of the Library of Congress." && img.Credit != "Courtesy of the National Archives." && img.Credit != "Courtesy of the John F. Kennedy Presidential Library and Museum." && img.Credit != "Courtesy of the National Park Service." && img.Credit != "Courtesy of the U.S. Geological Survey." && img.Credit != "Courtesy of the U.S. Senate." && img.Credit != "Courtesy of NASA." && img.Credit != "Courtesy of the White House." {
			return fmt.Errorf("image %s has unreviewed credit", img.ID)
		}
		for _, id := range img.Chapters {
			if !chapterIDs[id] {
				return fmt.Errorf("image %s: unknown chapter %s", img.ID, id)
			}
		}
		b, err := fs.ReadFile(files, "data/images/"+img.File)
		if err != nil {
			return fmt.Errorf("image %s: %w", img.ID, err)
		}
		size, format, err := image.DecodeConfig(bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("image %s: %w", img.ID, err)
		}
		if (format == "jpeg" && path.Ext(img.File) != ".jpg") || (format == "png" && path.Ext(img.File) != ".png") {
			return fmt.Errorf("image %s: extension does not match format", img.ID)
		}
		img.Width, img.Height = size.Width, size.Height
		imageIDs[img.ID] = img
		imageFiles[img.File] = true
	}
	entries, err := fs.ReadDir(files, "data/images")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "manifest.json" {
			continue
		}
		if !imageFiles[entry.Name()] {
			return fmt.Errorf("unmanifested image %s", entry.Name())
		}
	}
	for i := range c.Questions {
		c.Questions[i].Chapters = nil
		c.Questions[i].Library = nil
	}
	libraryIDs := map[string]bool{}
	usedImages := map[string]bool{}
	for _, doc := range c.Library {
		if libraryIDs[doc.ID] {
			return fmt.Errorf("duplicate library document %s", doc.ID)
		}
		libraryIDs[doc.ID] = true
	}
	for i := range c.Library {
		doc := &c.Library[i]
		seen := map[int]bool{}
		for _, id := range doc.Questions {
			if id < 1 || id > len(c.Questions) || seen[id] {
				return fmt.Errorf("library document %s: invalid/duplicate question %d", doc.ID, id)
			}
			seen[id] = true
			c.Questions[id-1].Library = append(c.Questions[id-1].Library, doc.ID)
		}
		seenImages := map[string]bool{}
		for j := range doc.Blocks {
			block := &doc.Blocks[j]
			if block.Kind != "image" {
				continue
			}
			img := imageIDs[block.ImageID]
			if img == nil || !slices.Contains(doc.Images, img.ID) || !slices.Contains(img.Library, doc.ID) {
				return fmt.Errorf("library document %s: unresolved image %s", doc.ID, block.ImageID)
			}
			if block.Page != img.Page || block.Text != img.Caption {
				return fmt.Errorf("image %s: source page/caption mismatch", img.ID)
			}
			block.Image = img
			seenImages[img.ID] = true
			usedImages[img.ID] = true
		}
		for _, id := range doc.Images {
			if !seenImages[id] {
				return fmt.Errorf("library document %s: unused image %s", doc.ID, id)
			}
		}
	}
	for i := range c.Chapters {
		ch := &c.Chapters[i]
		seen := map[int]bool{}
		if len(ch.Questions) == 0 {
			return fmt.Errorf("chapter %s has no questions", ch.ID)
		}
		for _, id := range ch.Questions {
			if id < 1 || id > len(c.Questions) || seen[id] {
				return fmt.Errorf("chapter %s: invalid/duplicate question %d", ch.ID, id)
			}
			seen[id] = true
			c.Questions[id-1].Chapters = append(c.Questions[id-1].Chapters, ch.ID)
		}
		seenImages := map[string]bool{}
		for j := range ch.Blocks {
			block := &ch.Blocks[j]
			if block.Kind != "image" {
				continue
			}
			img := imageIDs[block.ImageID]
			if img == nil || !slices.Contains(ch.Images, img.ID) || !slices.Contains(img.Chapters, ch.ID) {
				return fmt.Errorf("chapter %s: unresolved image %s", ch.ID, block.ImageID)
			}
			if block.Page != img.Page || block.Text != img.Caption {
				return fmt.Errorf("image %s: source page/caption mismatch", img.ID)
			}
			block.Image = img
			seenImages[img.ID] = true
			usedImages[img.ID] = true
		}
		listed := map[string]bool{}
		for _, id := range ch.Images {
			if !seenImages[id] || listed[id] {
				return fmt.Errorf("chapter %s: unused/duplicate image %s", ch.ID, id)
			}
			listed[id] = true
		}
	}
	for _, img := range c.Images {
		if !usedImages[img.ID] {
			return fmt.Errorf("unused image %s", img.ID)
		}
		for _, id := range img.Chapters {
			for _, ch := range c.Chapters {
				if ch.ID == id && !slices.Contains(ch.Images, img.ID) {
					return fmt.Errorf("image %s: missing reverse chapter link", img.ID)
				}
			}
		}
	}
	return nil
}

// ImageBytes reads only a manifest-approved runtime image, never raw extraction.
func (c *Catalog) ImageBytes(file string) ([]byte, error) {
	for _, img := range c.Images {
		if img.File == file {
			return data.ReadFile("data/images/" + file)
		}
	}
	return nil, fs.ErrNotExist
}
