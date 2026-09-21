package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Only these visually reviewed Study Guide images have approved printed credits.
// Keep this explicit allowlist: extraction order alone is not a licensing check.
func extractCurated(source, out string) error {
	tmp, err := os.MkdirTemp("", "n400-curated-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	for _, page := range []int{9, 13, 22, 26, 44, 46, 48, 49, 50, 53, 58, 59, 60, 62, 63, 64, 65, 67, 69, 73} {
		n := strconv.Itoa(page)
		if err := command("pdfimages", "-f", n, "-l", n, "-j", "-png", filepath.Join(source, "USCIS-2025-Civics-Test-Study-Guide.pdf"), filepath.Join(tmp, "p"+n)); err != nil {
			return err
		}
	}
	images := []struct{ extracted, name string }{
		{"p9-000.jpg", "ch01-signing.jpg"},
		{"p9-001.jpg", "ch01-constitution.jpg"},
		{"p13-002.png", "ch01-treaty.png"},
		{"p22-000.png", "ch02-oval-office.png"},
		{"p26-002.png", "ch03-voting.png"},
		{"p44-001.jpg", "ch07-jamestown-street.jpg"},
		{"p46-000.png", "ch07-pilgrims.png"},
		{"p48-000.jpg", "ch08-washington-command.jpg"},
		{"p48-001.jpg", "ch08-washington-princeton.jpg"},
		{"p49-000.jpg", "ch08-declaration.jpg"},
		{"p49-001.png", "ch08-trumbull.png"},
		{"p49-002.jpg", "ch08-independence-hall.jpg"},
		{"p50-000.jpg", "ch08-yorktown.jpg"},
		{"p53-001.jpg", "ch09-constitution.jpg"},
		{"p58-001.jpg", "ch10-cotton.jpg"},
		{"p59-000.png", "ch10-douglass.png"},
		{"p59-001.png", "ch10-anthony.png"},
		{"p60-002.jpg", "ch10-emancipation.jpg"},
		{"p62-001.jpg", "ch11-rubber-factories.jpg"},
		{"p63-000.jpg", "ch11-wwi-trench.jpg"},
		{"p63-001.jpg", "ch11-great-war-ends.jpg"},
		{"p63-002.jpg", "ch11-depression-breadline.jpg"},
		{"p63-003.png", "ch11-migrant-mother.png"},
		{"p64-000.png", "ch11-fdr-declaration.png"},
		{"p64-001.png", "ch11-pearl-harbor.png"},
		{"p65-000.png", "ch11-eisenhower.png"},
		{"p65-001.jpg", "ch11-moon-landing.jpg"},
		{"p67-000.jpg", "ch11-pentagon-flag.jpg"},
		{"p69-001.jpg", "ch12-statue-of-liberty.jpg"},
		{"p73-000.jpg", "ch12-washington-princeton.jpg"},
		{"p73-001.png", "ch12-lincoln.png"},
	}
	dir := filepath.Join(out, "images")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	for _, img := range images {
		b, err := os.ReadFile(filepath.Join(tmp, img.extracted))
		if err != nil {
			return fmt.Errorf("reading curated image %s: %w", img.name, err)
		}
		if err := os.WriteFile(filepath.Join(dir, img.name), b, 0644); err != nil {
			return fmt.Errorf("writing curated image %s: %w", img.name, err)
		}
	}
	return nil
}
