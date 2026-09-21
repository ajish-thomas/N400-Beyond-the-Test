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
	for _, page := range []int{9, 13, 22, 26, 44, 46, 48, 49, 50, 53, 58, 59, 60} {
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
