// Ingest is a development-only PDF extraction tool. Runtime builds need no PDFs.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"n400/internal/content"
)

func main() {
	source := flag.String("source", ".", "directory containing the four source PDFs")
	out := flag.String("out", "internal/content/data", "extraction output directory")
	images := flag.Bool("images", false, "extract unreviewed Study Guide images for local curation (never embedded)")
	curated := flag.Bool("curated-images", false, "extract the three reviewed Chapter 1 images")
	flag.Parse()
	if err := run(*source, *out, *images); err != nil {
		log.Fatal(err)
	}
	if *curated {
		if err := extractCurated(*source, *out); err != nil {
			log.Fatal(err)
		}
	}
}

func run(source, out string, images bool) error {
	raw := filepath.Join(out, "raw")
	if err := os.MkdirAll(raw, 0755); err != nil {
		return err
	}
	files := []struct{ pdf, text string }{
		{"2025-Civics-Test-128-Questions-and-Answers.pdf", "questions.txt"},
		{"USCIS-2025-Civics-Test-Study-Guide.pdf", "study-guide.txt"},
		{"CitizensAlmanac-M-76.pdf", "almanac.txt"},
		{"DOI-Constitution-M-654.pdf", "constitution.txt"},
	}
	for _, file := range files {
		if err := command("pdftotext", "-layout", filepath.Join(source, file.pdf), filepath.Join(raw, file.text)); err != nil {
			return err
		}
	}
	b, err := os.ReadFile(filepath.Join(raw, "questions.txt"))
	if err != nil {
		return err
	}
	counts, err := content.RequiredCounts()
	if err != nil {
		return err
	}
	questions, err := content.ParseQuestions(strings.NewReader(string(b)), counts)
	if err != nil {
		return err
	}
	b, err = json.MarshalIndent(questions, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "questions.json"), append(b, '\n'), 0644); err != nil {
		return err
	}
	if images {
		dir := filepath.Join(out, "images", "_extracted")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		// Never extract Almanac images. Study Guide output is unreviewed and may
		// include AP material; it must not be published or embedded as-is.
		if err := command("pdfimages", "-j", filepath.Join(source, files[1].pdf), filepath.Join(dir, "study")); err != nil {
			return err
		}
	}
	log.Printf("extracted four texts and validated %d questions", len(questions))
	return nil
}

func command(name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	b, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w: %s", name, err, b)
	}
	return nil
}
