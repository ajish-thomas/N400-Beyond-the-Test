package content

import "fmt"

var starredIDs = map[int]bool{2: true, 7: true, 12: true, 20: true, 30: true, 36: true, 38: true, 39: true, 44: true, 52: true, 61: true, 66: true, 74: true, 78: true, 86: true, 94: true, 113: true, 115: true, 121: true, 126: true}

func Validate(questions []Question) error {
	if len(questions) != 128 {
		return fmt.Errorf("content: want 128 questions, got %d", len(questions))
	}
	counts, err := RequiredCounts()
	if err != nil {
		return err
	}
	for i, q := range questions {
		if q.ID != i+1 {
			return fmt.Errorf("content: expected question %d, got %d", i+1, q.ID)
		}
		want := 1
		if n, ok := counts[q.ID]; ok {
			want = n
		}
		if q.RequiredCount != want || want < 1 || want > len(q.Answers) {
			return fmt.Errorf("question %d: invalid answer cardinality", q.ID)
		}
		if q.Prompt == "" || subsections[q.Subsection] != q.Section || q.Section == "" {
			return fmt.Errorf("question %d: missing or invalid prompt/section", q.ID)
		}
		if q.Is6520 != starredIDs[q.ID] {
			return fmt.Errorf("question %d: incorrect 65/20 marking", q.ID)
		}
		if q.AnswerKind != kindFor(q.ID) {
			return fmt.Errorf("question %d: incorrect answer kind", q.ID)
		}
		for _, a := range q.Answers {
			parsed, err := ParseAnswer(a.Text)
			if err != nil {
				return fmt.Errorf("question %d: %w", q.ID, err)
			}
			if a != parsed || a.Core == "" || a.Full == "" {
				return fmt.Errorf("question %d: invalid answer fields", q.ID)
			}
		}
	}
	return nil
}
