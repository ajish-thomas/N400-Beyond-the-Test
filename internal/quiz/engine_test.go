package quiz

import (
	"math/rand"
	"slices"
	"testing"

	"n400/internal/content"
)

func TestSessionSelectionIsSeededAndUnique(t *testing.T) {
	questions, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	a, err := NewSession(questions, Official, rand.New(rand.NewSource(42)))
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewSession(questions, Official, rand.New(rand.NewSource(42)))
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Questions) != 20 || a.PassAt != 12 {
		t.Fatalf("official session: %+v", a)
	}
	seen := map[int]bool{}
	for i, q := range a.Questions {
		if seen[q.ID] || q.ID != b.Questions[i].ID {
			t.Fatal("selection must be unique and deterministic")
		}
		seen[q.ID] = true
	}
	starred, err := NewSession(questions, Study6520, rand.New(rand.NewSource(7)))
	if err != nil {
		t.Fatal(err)
	}
	if len(starred.Questions) != 10 || starred.PassAt != 6 {
		t.Fatalf("65/20 session: %+v", starred)
	}
	for _, q := range starred.Questions {
		if !q.Is6520 {
			t.Fatalf("non-starred Q%d selected", q.ID)
		}
	}
	if _, err := NewSession(questions, "bad", rand.New(rand.NewSource(1))); err == nil {
		t.Fatal("accepted invalid mode")
	}
}

func TestSessionStopsAtPassAndFailureBoundaries(t *testing.T) {
	questions := make([]content.Question, 20)
	for i := range questions {
		questions[i] = content.Question{ID: i + 1}
	}
	pass := Session{Questions: slices.Clone(questions), PassAt: 12}
	for i := 0; i < 12; i++ {
		if err := pass.Record(true); err != nil {
			t.Fatal(err)
		}
	}
	if !pass.Complete || !pass.Passed || pass.Answered != 12 {
		t.Fatalf("pass boundary: %+v", pass)
	}
	fail := Session{Questions: slices.Clone(questions), PassAt: 12}
	for i := 0; i < 9; i++ {
		if err := fail.Record(false); err != nil {
			t.Fatal(err)
		}
	}
	if !fail.Complete || fail.Passed || fail.Answered != 9 {
		t.Fatalf("failure boundary: %+v", fail)
	}
}
