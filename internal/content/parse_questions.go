package content

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var questionLine = regexp.MustCompile(`^(\d+)\.\s+(.+)$`)
var footerLine = regexp.MustCompile(`^\d+ of 19\s+uscis\.gov/citizenship$`)

var subsections = map[string]string{
	"A: Principles of American Government": "AMERICAN GOVERNMENT",
	"B: System of Government":              "AMERICAN GOVERNMENT",
	"C: Rights and Responsibilities":       "AMERICAN GOVERNMENT",
	"A: Colonial Period and Independence":  "AMERICAN HISTORY",
	"B: 1800s":                             "AMERICAN HISTORY",
	"C: Recent American History and Other Important Historical Information": "AMERICAN HISTORY",
	"A: Symbols":  "SYMBOLS AND HOLIDAYS",
	"B: Holidays": "SYMBOLS AND HOLIDAYS",
}

func kindFor(id int) Kind {
	switch id {
	case 23, 29, 61, 62:
		return StateSpecific
	case 30, 38, 39, 57:
		return CurrentOfficial
	default:
		return Fixed
	}
}

// ParseQuestions accepts pdftotext -layout output, never a PDF.
func ParseQuestions(r io.Reader, counts map[int]int) ([]Question, error) {
	var questions []Question
	var current *Question
	section, subsection := "", ""
	scanner := bufio.NewScanner(r)
	for line := 1; scanner.Scan(); line++ {
		s := reflow(scanner.Text())
		if s == "" || footerLine.MatchString(s) {
			continue
		}
		switch s {
		case "AMERICAN GOVERNMENT", "AMERICAN HISTORY", "SYMBOLS AND HOLIDAYS":
			section, subsection = s, ""
			continue
		}
		if parent, ok := subsections[s]; ok {
			if parent != section {
				return nil, fmt.Errorf("line %d: subsection in wrong section", line)
			}
			subsection = s
			continue
		}
		if match := questionLine.FindStringSubmatch(s); match != nil {
			id, err := strconv.Atoi(match[1])
			if err != nil {
				return nil, fmt.Errorf("line %d: question ID: %w", line, err)
			}
			if id != len(questions)+1 || subsection == "" {
				return nil, fmt.Errorf("line %d: unexpected question %d", line, id)
			}
			count := 1
			if n, ok := counts[id]; ok {
				count = n
			}
			questions = append(questions, Question{ID: id, Section: section, Subsection: subsection, Prompt: match[2], RequiredCount: count, AnswerKind: kindFor(id)})
			current = &questions[len(questions)-1]
			continue
		}
		if current == nil {
			if section != "" {
				return nil, fmt.Errorf("line %d: unexpected text before first question: %q", line, s)
			}
			continue // introductory page is retained in the raw source
		}
		if s == "*" {
			current.Is6520 = true
			continue
		}
		if strings.HasPrefix(s, "•") {
			current.Answers = append(current.Answers, Answer{Text: strings.TrimSpace(strings.TrimPrefix(s, "•"))})
		} else if s == "For a complete list of tribes, please visit bia.gov." && current.ID == 117 {
			current.Note = s
		} else if len(current.Answers) == 0 {
			current.Prompt += " " + s
		} else {
			current.Answers[len(current.Answers)-1].Text += " " + s
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading questions: %w", err)
	}
	for i := range questions {
		q := &questions[i]
		if strings.HasSuffix(q.Prompt, "*") {
			q.Is6520 = true
			q.Prompt = strings.TrimSpace(strings.TrimSuffix(q.Prompt, "*"))
		}
		for j := range q.Answers {
			a, err := ParseAnswer(q.Answers[j].Text)
			if err != nil {
				return nil, fmt.Errorf("parsing question %d answer %d: %w", q.ID, j+1, err)
			}
			q.Answers[j] = a
		}
	}
	if err := Validate(questions); err != nil {
		return nil, err
	}
	return questions, nil
}
