package flashcard

import (
	"testing"
	"time"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestReviewPromotesAndDemotes(t *testing.T) {
	now := time.Date(2026, time.September, 20, 15, 30, 0, 0, time.FixedZone("PDT", -7*60*60))
	scheduler := NewScheduler(fixedClock{now: now})

	card := scheduler.Review(Card{}, true)
	if card.Box != 2 || !card.Due.Equal(time.Date(2026, time.September, 22, 0, 0, 0, 0, now.Location())) {
		t.Fatalf("new correct card = %#v", card)
	}
	card = scheduler.Review(card, true)
	if card.Box != 3 || card.Due.Day() != 24 {
		t.Fatalf("second correct card = %#v", card)
	}
	card = scheduler.Review(card, false)
	if card.Box != 1 || card.Due.Day() != 21 {
		t.Fatalf("wrong card = %#v", card)
	}
}

func TestReviewCapsAtFinalBox(t *testing.T) {
	now := time.Date(2026, time.September, 20, 12, 0, 0, 0, time.UTC)
	scheduler := NewScheduler(fixedClock{now: now})
	card := scheduler.Review(Card{Box: MaxBox}, true)
	if card.Box != MaxBox || !card.Due.Equal(time.Date(2026, time.October, 6, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("final box card = %#v", card)
	}
}

func TestDueAcrossDayBoundary(t *testing.T) {
	location := time.FixedZone("PDT", -7*60*60)
	clock := fixedClock{now: time.Date(2026, time.September, 20, 23, 59, 0, 0, location)}
	scheduler := NewScheduler(clock)
	if !scheduler.Due(Card{}, false) {
		t.Fatal("unseen card should be due")
	}
	card := scheduler.Review(Card{}, false)
	if scheduler.Due(card, true) {
		t.Fatal("tomorrow's card should not be due before midnight")
	}
	scheduler = NewScheduler(fixedClock{now: time.Date(2026, time.September, 21, 0, 1, 0, 0, location)})
	if !scheduler.Due(card, true) {
		t.Fatal("card should be due after its calendar day begins")
	}
}
