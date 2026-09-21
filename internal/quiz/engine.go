package quiz

import (
	"fmt"
	"math/rand"

	"n400/internal/content"
)

type Mode string

const (
	Official  Mode = "official"
	Study6520 Mode = "6520"
)

type Session struct {
	Questions []content.Question
	PassAt    int
	Correct   int
	Answered  int
	Complete  bool
	Passed    bool
}

// NewSession draws a no-repeat mock test using only the supplied RNG, making
// selection reproducible in tests and independent of the system clock.
func NewSession(questions []content.Question, mode Mode, rng *rand.Rand) (Session, error) {
	if rng == nil {
		return Session{}, fmt.Errorf("quiz RNG is required")
	}
	count, passAt := 20, 12
	pool := append([]content.Question(nil), questions...)
	if mode == Study6520 {
		count, passAt = 10, 6
		pool = make([]content.Question, 0, 20)
		for _, q := range questions {
			if q.Is6520 {
				pool = append(pool, q)
			}
		}
	} else if mode != Official {
		return Session{}, fmt.Errorf("unknown quiz mode %q", mode)
	}
	if len(pool) < count {
		return Session{}, fmt.Errorf("quiz mode %s needs %d questions, got %d", mode, count, len(pool))
	}
	rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	return Session{Questions: pool[:count], PassAt: passAt}, nil
}

// Record advances a session and stops as soon as passing or failing is certain.
func (s *Session) Record(correct bool) error {
	if s.Complete {
		return fmt.Errorf("quiz is already complete")
	}
	if s.Answered >= len(s.Questions) {
		return fmt.Errorf("quiz has no remaining questions")
	}
	s.Answered++
	if correct {
		s.Correct++
	}
	remaining := len(s.Questions) - s.Answered
	switch {
	case s.Correct >= s.PassAt:
		s.Complete, s.Passed = true, true
	case s.Correct+remaining < s.PassAt:
		s.Complete, s.Passed = true, false
	}
	return nil
}
