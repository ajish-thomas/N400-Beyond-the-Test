// Package content loads and validates the official USCIS question corpus.
package content

import (
	"embed"
	"encoding/json"
	"fmt"
)

type Kind string

const (
	Fixed           Kind = "fixed"
	StateSpecific   Kind = "state_specific"
	CurrentOfficial Kind = "current_official"
)

type Answer struct {
	Text     string
	Core     string
	Full     string
	Guidance string
}

type Question struct {
	ID            int
	Section       string
	Subsection    string
	Prompt        string
	Answers       []Answer
	RequiredCount int
	Is6520        bool
	AnswerKind    Kind
	Chapters      []string
	Note          string
}

func (q Question) Changing() bool { return q.AnswerKind != Fixed }

// Only curated runtime data is embedded; raw source and unreviewed images are not.
//
//go:embed data/required_counts.json data/questions.json
var data embed.FS

func RequiredCounts() (map[int]int, error) {
	b, err := data.ReadFile("data/required_counts.json")
	if err != nil {
		return nil, err
	}
	var counts map[int]int
	if err := json.Unmarshal(b, &counts); err != nil {
		return nil, fmt.Errorf("decoding required counts: %w", err)
	}
	return counts, nil
}

func Load() ([]Question, error) {
	b, err := data.ReadFile("data/questions.json")
	if err != nil {
		return nil, err
	}
	var questions []Question
	if err := json.Unmarshal(b, &questions); err != nil {
		return nil, fmt.Errorf("decoding questions: %w", err)
	}
	if err := Validate(questions); err != nil {
		return nil, err
	}
	return questions, nil
}
