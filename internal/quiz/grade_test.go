package quiz

import (
	"strings"
	"testing"

	"n400/internal/content"
)

func TestEveryFixedOfficialAnswerMatchesAsAnItem(t *testing.T) {
	questions, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range questions {
		if q.Changing() {
			if got := Grade(q, q.Answers[0].Text); !got.Unavailable {
				t.Errorf("Q%d changing instruction should not grade as an answer", q.ID)
			}
			continue
		}
		for _, answer := range q.Answers {
			got := Grade(q, answer.Text)
			if len(got.Matched) != 1 || got.Matched[0] != answer.Text {
				t.Errorf("Q%d %q: matched %#v", q.ID, answer.Text, got.Matched)
			}
			if q.RequiredCount == 1 && !got.Correct {
				t.Errorf("Q%d %q should pass: %+v", q.ID, answer.Text, got)
			}
			if q.RequiredCount > 1 && got.Correct {
				t.Errorf("Q%d one item should not satisfy %d answers", q.ID, q.RequiredCount)
			}
		}
	}
}

func TestEnumerationCardinalityAndDistinctness(t *testing.T) {
	questions, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{10, 48, 65, 67, 69, 81, 126} {
		q := questions[id-1]
		var values []string
		for i := 0; i < q.RequiredCount; i++ {
			values = append(values, q.Answers[i].Text)
		}
		if got := Grade(q, strings.Join(values, ", ")); !got.Correct || len(got.Matched) != q.RequiredCount {
			t.Errorf("Q%d distinct answers: %+v", id, got)
		}
		if got := Grade(q, strings.Join(values[:q.RequiredCount-1], "; ")); got.Correct || got.Missing != 1 {
			t.Errorf("Q%d N-1 answers: %+v", id, got)
		}
		duplicates := make([]string, q.RequiredCount)
		for i := range duplicates {
			duplicates[i] = q.Answers[0].Text
		}
		if got := Grade(q, strings.Join(duplicates, "\n")); got.Correct || len(got.Matched) != 1 {
			t.Errorf("Q%d duplicates: %+v", id, got)
		}
	}
	q := questions[68] // Q69
	if got := Grade(q, "Vote and run for office"); !got.Correct {
		t.Fatalf("mixed separator: %+v", got)
	}
}

func TestNormalizationAndGuidance(t *testing.T) {
	questions, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		id   int
		text string
	}{
		{2, "Constitution"},
		{2, "the U.S. Constitution"},
		{7, "27"},
		{7, "twenty-seven"},
		{81, "New York; New Jersey; Pennsylvania; Delaware; Maryland"},
		{16, "legislative, exectuive and judiciary"},
		{16, "the judicial branch, executive branch, and legislative branch"},
		{16, "the branch that makes federal laws, the branch that enforces federal laws, and the branch that reviews federal laws"},
		{110, "stop communism"},
		{111, "stop communism"},
		{112, "end racial discrimination"},
	} {
		if got := Grade(questions[tc.id-1], tc.text); !got.Correct {
			t.Errorf("Q%d %q: %+v", tc.id, tc.text, got)
		}
	}
	if got := Grade(questions[22], "D.C. has no U.S. senators"); !got.Unavailable || got.Correct {
		t.Fatalf("guidance must not become an answer: %+v", got)
	}
	if Normalize("The café — U.S. 27") != "cafe united states twenty seven" {
		t.Fatalf("unexpected normalization: %q", Normalize("The café — U.S. 27"))
	}
}

func TestSourceReviewedSemanticRules(t *testing.T) {
	questions, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range semanticRules {
		q := questions[rule.QuestionID-1]
		if rule.Source == "" || rule.AnswerIndex < 0 || rule.AnswerIndex >= len(q.Answers) {
			t.Fatalf("invalid semantic rule: %+v", rule)
		}
		got := Grade(q, strings.Join(rule.Terms, " "))
		if !got.Correct || len(got.Matched) != 1 || got.Matched[0] != q.Answers[rule.AnswerIndex].Text {
			t.Errorf("Q%d %q: %+v", q.ID, strings.Join(rule.Terms, " "), got)
		}
	}
}

func TestGradeWithResolvedAnswer(t *testing.T) {
	question := content.Question{ID: 38, AnswerKind: content.CurrentOfficial, RequiredCount: 1, Answers: []content.Answer{{Text: "Answers will vary."}}}
	if result := GradeWithResolvedAnswer(question, "JD Vance", "JD Vance"); !result.Correct || result.Unavailable {
		t.Fatalf("resolved answer = %#v", result)
	}
	if result := GradeWithResolvedAnswer(question, "JD Vance", ""); !result.Unavailable {
		t.Fatalf("missing resolution = %#v", result)
	}
}
