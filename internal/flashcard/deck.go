package flashcard

import "n400/internal/content"

// Filter narrows a deck without changing the official question order.
type Filter struct {
	Section  string
	Chapter  string
	Only6520 bool
	Missed   bool
}

// Build selects questions for a flashcard deck. Missed is keyed by question ID
// and comes from the local progress store once persistence is enabled.
func Build(questions []content.Question, filter Filter, missed map[int]bool) []content.Question {
	deck := make([]content.Question, 0, len(questions))
	for _, question := range questions {
		if filter.Section != "" && question.Section != filter.Section {
			continue
		}
		if filter.Chapter != "" && !contains(question.Chapters, filter.Chapter) {
			continue
		}
		if filter.Only6520 && !question.Is6520 {
			continue
		}
		if filter.Missed && !missed[question.ID] {
			continue
		}
		deck = append(deck, question)
	}
	return deck
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
