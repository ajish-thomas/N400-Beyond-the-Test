package content

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestSourceParseGolden(t *testing.T) {
	raw, err := os.ReadFile("data/raw/questions.txt")
	if err != nil {
		t.Fatal(err)
	}
	counts, err := RequiredCounts()
	if err != nil {
		t.Fatal(err)
	}
	questions, err := ParseQuestions(bytes.NewReader(raw), counts)
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.MarshalIndent(questions, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("data/questions.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(append(got, '\n'), want) {
		t.Fatal("full parse differs from reviewed questions.json; inspect source before updating")
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(questions, loaded) {
		t.Fatal("embedded questions differ from raw parse")
	}
	// Independent source checks protect the initial golden from common omissions.
	if len(questions[47].Answers) != 22 || len(questions[66].Answers) != 6 {
		t.Fatal("source answer menus truncated")
	}
	if questions[116].Note != "For a complete list of tribes, please visit bia.gov." {
		t.Fatal("standalone source note lost")
	}
	if questions[116].Answers[24].Text != "Tuscarora" {
		t.Fatal("standalone note merged into answer")
	}
	if !strings.Contains(questions[61].Answers[0].Guidance, "capital of the territory.") {
		t.Fatal("wrapped guidance lost")
	}
	if questions[96].Prompt != "What amendment says all persons born or naturalized in the United States, and subject to the jurisdiction thereof, are U.S. citizens?" {
		t.Fatal("wrapped prompt corrupted")
	}
	// There is one parsed answer per printed bullet, including wrapped bullets.
	bulletCount := strings.Count(string(raw), "•")
	answerCount := 0
	for _, q := range questions {
		answerCount += len(q.Answers)
	}
	if answerCount != bulletCount {
		t.Fatalf("got %d answers for %d printed bullets", answerCount, bulletCount)
	}
}

func TestRequiredCountsAndCardinalityGuard(t *testing.T) {
	counts, err := RequiredCounts()
	if err != nil {
		t.Fatal(err)
	}
	want := map[int]int{10: 2, 48: 2, 65: 3, 67: 2, 69: 2, 81: 5, 126: 3}
	if !reflect.DeepEqual(counts, want) {
		t.Fatalf("cardinality contract changed: %v", counts)
	}
	questions, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	// Reviewed wording about existing counts, not requests for separate answers.
	exceptions := map[int]bool{15: true, 16: true, 19: true, 28: true, 37: true, 63: true}
	// Dates, ordinals and counts already expressed as digits occur in descriptive
	// prompts; this guard targets words used by the source's enumeration prompts.
	words := regexp.MustCompile(`(?i)\b(two|three|four|five|six|seven|eight|nine|ten)\b`)
	for _, q := range questions {
		if words.MatchString(q.Prompt) && counts[q.ID] == 0 && !exceptions[q.ID] {
			t.Errorf("Q%d: unreviewed cardinality wording: %s", q.ID, q.Prompt)
		}
	}
}

func TestParseAnswer(t *testing.T) {
	for _, tc := range []struct{ text, core, full, guidance string }{
		{"(U.S.) Constitution", "Constitution", "U.S. Constitution", ""},
		{"Serve (help, do important work for) the nation (if needed)", "Serve the nation", "Serve help, do important work for the nation if needed", ""},
		{"Presidents Day (Washington’s Birthday)", "Presidents Day", "Presidents Day Washington’s Birthday", ""},
		{"Secretary of War (Defense)", "Secretary of War", "Secretary of War Defense", ""},
		{"Answers will vary. [D.C. (or the territory) has no senators.]", "Answers will vary.", "Answers will vary.", "D.C. (or the territory) has no senators."},
		{"Liberty", "Liberty", "Liberty", ""},
	} {
		t.Run(tc.text, func(t *testing.T) {
			got, err := ParseAnswer(tc.text)
			if err != nil {
				t.Fatal(err)
			}
			want := Answer{tc.text, tc.core, tc.full, tc.guidance}
			if got != want {
				t.Fatalf("got %#v, want %#v", got, want)
			}
		})
	}
	for _, text := range []string{"(open", "close)", "[open", "close]", "[nested [text]]"} {
		if _, err := ParseAnswer(text); err == nil {
			t.Errorf("accepted malformed notation: %q", text)
		}
	}
}

func TestParserFailuresAndDetachedStar(t *testing.T) {
	raw, err := os.ReadFile("data/raw/questions.txt")
	if err != nil {
		t.Fatal(err)
	}
	counts, err := RequiredCounts()
	if err != nil {
		t.Fatal(err)
	}
	original := string(raw)
	star := regexp.MustCompile(`(2\. What is the supreme law of the land\?)\s+\*`)
	detached := star.ReplaceAllString(original, "$1\n*\n")
	if _, err := ParseQuestions(strings.NewReader(detached), counts); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{
		strings.Replace(original, "2. What is the supreme", "3. What is the supreme", 1),
		strings.Replace(original, "(U.S.) Constitution", "(U.S. Constitution", 1),
		original[:len(original)/2],
	} {
		if _, err := ParseQuestions(strings.NewReader(bad), counts); err == nil {
			t.Error("accepted damaged source")
		}
	}
}

func TestValidationRejectsCorruption(t *testing.T) {
	for _, mutate := range []func([]Question){
		func(q []Question) { q[0].ID = 2 },
		func(q []Question) { q[1].Is6520 = false },
		func(q []Question) { q[9].RequiredCount = 1 },
		func(q []Question) { q[22].AnswerKind = Fixed },
		func(q []Question) { q[0].Subsection = "unknown" },
		func(q []Question) { q[0].Answers[0].Core = "invented" },
	} {
		q, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		mutate(q)
		if err := Validate(q); err == nil {
			t.Error("accepted corrupt content")
		}
	}
}
