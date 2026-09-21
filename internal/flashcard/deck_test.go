package flashcard

import (
	"reflect"
	"testing"

	"n400/internal/content"
)

func TestBuildAppliesDeckFilters(t *testing.T) {
	questions := []content.Question{
		{ID: 1, Section: "Government", Chapters: []string{"constitution"}, Is6520: true},
		{ID: 2, Section: "Government", Chapters: []string{"executive"}},
		{ID: 3, Section: "History", Chapters: []string{"revolution"}, Is6520: true},
	}
	got := Build(questions, Filter{Section: "Government", Only6520: true}, nil)
	if ids := questionIDs(got); !reflect.DeepEqual(ids, []int{1}) {
		t.Fatalf("filtered IDs = %v", ids)
	}
	got = Build(questions, Filter{Chapter: "revolution", Missed: true}, map[int]bool{3: true})
	if ids := questionIDs(got); !reflect.DeepEqual(ids, []int{3}) {
		t.Fatalf("missed chapter IDs = %v", ids)
	}
}

func questionIDs(questions []content.Question) []int {
	ids := make([]int, len(questions))
	for i, question := range questions {
		ids[i] = question.ID
	}
	return ids
}
