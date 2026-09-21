// Package flashcard implements local, deterministic spaced-repetition rules.
package flashcard

import (
	"time"
)

const (
	MinBox = 1
	MaxBox = 5
)

var intervals = [...]int{0, 1, 2, 4, 8, 16}

// Clock makes scheduling independent of the wall clock in tests and callers.
type Clock interface {
	Now() time.Time
}

// SystemClock is the production clock. Tests provide their own Clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

// Card is the persisted scheduling state for one question. Due is stored as a
// calendar day: the card becomes available at the beginning of that day.
type Card struct {
	Box int       `json:"box"`
	Due time.Time `json:"due"`
}

// Scheduler applies the five-box Leitner schedule. A Scheduler always has an
// injected clock so callers can make review behavior reproducible.
type Scheduler struct {
	clock Clock
}

func NewScheduler(clock Clock) Scheduler {
	if clock == nil {
		panic("flashcard clock is required")
	}
	return Scheduler{clock: clock}
}

// Review returns the next state after an answer. A new card begins in box one;
// a correct answer promotes it, while an incorrect answer returns it to box one.
func (s Scheduler) Review(previous Card, correct bool) Card {
	box := previous.Box
	if box < MinBox || box > MaxBox {
		box = MinBox
	}
	if correct && box < MaxBox {
		box++
	}
	if !correct {
		box = MinBox
	}
	now := calendarDay(s.clock.Now())
	return Card{Box: box, Due: now.AddDate(0, 0, intervals[box])}
}

// Due reports whether a persisted card is ready on the scheduler's current
// calendar day. Unseen cards are due immediately.
func (s Scheduler) Due(card Card, seen bool) bool {
	return !seen || !calendarDay(card.Due).After(calendarDay(s.clock.Now()))
}

func calendarDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
