// Package quiz implements advisory grading and test-session rules.
package quiz

import (
	"strconv"
	"strings"
	"unicode"

	"n400/internal/content"
)

// Result explains what the grader recognized. Correct is never a substitute
// for the user's self-check: the USCIS interview is oral and officer-judged.
type Result struct {
	Correct     bool
	Unavailable bool
	Required    int
	Matched     []string
	Missing     int
	Message     string
}

// Grade recognizes fixed official answers generously. Changing answers have no
// bundled factual value, so callers must supply their resolved value later;
// accepting the PDF's “Answers will vary” instruction would be misleading.
func Grade(q content.Question, response string) Result {
	r := Result{Required: q.RequiredCount}
	if q.Changing() {
		r.Unavailable = true
		r.Message = "This changing answer needs your current location or official data. You can still mark it right yourself."
		return r
	}

	matched := map[int]bool{}
	normalized := semanticInput(q, Normalize(response))
	if q.RequiredCount == 1 {
		if index := semanticRuleMatch(q, normalized); index >= 0 {
			matched[index] = true
		} else if index := matchAnswer(normalized, q.Answers); index >= 0 {
			matched[index] = true
		}
	} else {
		// Match each menu item against the whole response. This is equivalent to
		// accepting comma/semicolon/newline/“and” separators for this corpus,
		// while also preserving official answers that themselves contain commas
		// or “and” (for example Martin Luther King, Jr. Day).
		for i, answer := range q.Answers {
			if answerMatches(normalized, answer) {
				matched[i] = true
			}
		}
	}
	for i, answer := range q.Answers {
		if matched[i] {
			r.Matched = append(r.Matched, answer.Text)
		}
	}
	r.Missing = q.RequiredCount - len(r.Matched)
	if r.Missing < 0 {
		r.Missing = 0
	}
	r.Correct = r.Missing == 0
	switch {
	case r.Correct && q.RequiredCount == 1:
		r.Message = "That matches an official answer."
	case r.Correct:
		r.Message = "You named the required number of distinct official answers."
	case len(r.Matched) == 0:
		r.Message = "I did not recognize that wording. The acceptable official answers are shown below. Use your own judgment."
	default:
		r.Message = "Matched " + strconv.Itoa(len(r.Matched)) + " of " + strconv.Itoa(q.RequiredCount) + " required; name " + strconv.Itoa(r.Missing) + " more distinct answer"
		if r.Missing != 1 {
			r.Message += "s"
		}
		r.Message += "."
	}
	return r
}

// GradeWithResolvedAnswer grades a changing prompt only after its answer has
// been resolved locally. It never treats the PDF's “Answers will vary”
// instruction as an answer.
func GradeWithResolvedAnswer(q content.Question, response, resolved string) Result {
	return GradeWithResolvedAnswers(q, response, []string{resolved})
}

// GradeWithResolvedAnswers accepts every factual answer resolved for a
// changing prompt, such as either of a state's two senators.
func GradeWithResolvedAnswers(q content.Question, response string, resolved []string) Result {
	answers := make([]content.Answer, 0, len(resolved))
	for _, answer := range resolved {
		if answer = strings.TrimSpace(answer); answer != "" {
			answers = append(answers, content.Answer{Text: answer, Core: answer, Full: answer})
		}
	}
	if len(answers) == 0 {
		return Grade(q, response)
	}
	q.AnswerKind = content.Fixed
	q.RequiredCount = 1
	q.Answers = answers
	return Grade(q, response)
}

// semanticInput adds only reviewed concepts stated in the official source
// material. It is intentionally small and question-specific: a general model
// should not decide that two nearby civics facts mean the same thing.
func semanticInput(q content.Question, input string) string {
	words := strings.Fields(input)
	var concepts []string
	if q.ID == 16 && hasMeaningfulWords(words, "make", "federal", "law") {
		concepts = append(concepts, "legislative")
	}
	if q.ID == 16 && hasMeaningfulWords(words, "enforce", "federal", "law") {
		concepts = append(concepts, "executive")
	}
	if q.ID == 16 && hasMeaningfulWords(words, "review", "federal", "law") {
		concepts = append(concepts, "judicial")
	}
	return strings.TrimSpace(input + " " + strings.Join(concepts, " "))
}

func semanticRuleMatch(q content.Question, input string) int {
	for _, rule := range semanticRules {
		if rule.QuestionID == q.ID && rule.AnswerIndex < len(q.Answers) && hasMeaningfulWords(strings.Fields(input), rule.Terms...) {
			return rule.AnswerIndex
		}
	}
	return -1
}

func hasMeaningfulWords(words []string, needed ...string) bool {
	for _, target := range needed {
		found := false
		for _, word := range words {
			if wordRoot(word) == target {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func matchAnswer(fragment string, answers []content.Answer) int {
	input := Normalize(fragment)
	if input == "" {
		return -1
	}
	// Preserve the self-match contract for every printed bullet. Text may carry
	// reader guidance, but it is accepted only when the complete bullet itself
	// was submitted; guidance alone never reaches this branch.
	for i, answer := range answers {
		if input == Normalize(answer.Text) {
			return i
		}
	}
	for i, answer := range answers {
		if input == Normalize(answer.Core) || input == Normalize(answer.Full) {
			return i
		}
	}
	best, bestLength := -1, -1
	for i, answer := range answers {
		if !answerMatches(input, answer) {
			continue
		}
		length := len(strings.Fields(Normalize(answer.Full)))
		if length > bestLength {
			best, bestLength = i, length
		}
	}
	if best < 0 && len(answers) == 1 && closeMeaningfulWords(strings.Fields(input), strings.Fields(Normalize(answers[0].Core))) {
		return 0
	}
	return best
}

func answerMatches(input string, answer content.Answer) bool {
	for _, acceptable := range []string{Normalize(answer.Core), Normalize(answer.Full)} {
		if acceptable != "" && containsTokens(input, acceptable) {
			return true
		}
	}
	return false
}

var fillerWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
	"by": true, "for": true, "from": true, "in": true, "is": true, "it": true,
	"of": true, "on": true, "or": true, "the": true, "to": true, "was": true,
	"with": true,
}

// sameMeaningfulWords accepts a spoken reordering such as “the legislative
// branch, executive branch, and judiciary” for “legislative, executive, and
// judicial.” It deliberately requires every meaningful word from the official
// answer, so it does not turn a vague topical response into a correct answer.
func sameMeaningfulWords(inputWords, answerWords []string) bool {
	meaningful := make([]string, 0, len(answerWords))
	for _, word := range answerWords {
		if !fillerWords[word] {
			meaningful = append(meaningful, wordRoot(word))
		}
	}
	if len(meaningful) < 2 {
		return false
	}
	for _, target := range meaningful {
		found := false
		for _, input := range inputWords {
			if equivalentWord(wordRoot(input), target) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// closeMeaningfulWords is reserved for a question with exactly one official
// answer. It accepts an oral response that supplies all but one meaningful
// official term, provided it still supplies at least two. For example, “stop
// communism” identifies the sole Korean-War answer “stop the spread of
// communism.” Multi-answer menus never use this relaxed rule.
func closeMeaningfulWords(inputWords, answerWords []string) bool {
	meaningful := make([]string, 0, len(answerWords))
	for _, word := range answerWords {
		if !fillerWords[word] {
			meaningful = append(meaningful, wordRoot(word))
		}
	}
	if len(meaningful) < 3 {
		return false
	}
	matched := 0
	for _, target := range meaningful {
		for _, input := range inputWords {
			if equivalentWord(wordRoot(input), target) {
				matched++
				break
			}
		}
	}
	return matched >= 2 && matched >= len(meaningful)-1
}

func wordRoot(word string) string {
	switch {
	case strings.HasSuffix(word, "ies") && len(word) > 4:
		return strings.TrimSuffix(word, "ies") + "y"
	case strings.HasSuffix(word, "ches"), strings.HasSuffix(word, "shes"), strings.HasSuffix(word, "sses"), strings.HasSuffix(word, "xes"), strings.HasSuffix(word, "zes"):
		return strings.TrimSuffix(word, "es")
	case strings.HasSuffix(word, "ss"):
		return word
	case strings.HasSuffix(word, "s") && len(word) > 3:
		return strings.TrimSuffix(word, "s")
	default:
		return word
	}
}

func containsTokens(input, acceptable string) bool {
	needle := " " + acceptable + " "
	if strings.Contains(" "+input+" ", needle) {
		return true
	}
	inputWords, answerWords := strings.Fields(input), strings.Fields(acceptable)
	if len(answerWords) == 0 || len(inputWords) < len(answerWords) {
		return false
	}
	for start := 0; start+len(answerWords) <= len(inputWords); start++ {
		matches := true
		for i, answerWord := range answerWords {
			if !equivalentWord(inputWords[start+i], answerWord) {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return sameMeaningfulWords(inputWords, answerWords)
}

func equivalentWord(input, answer string) bool {
	if input == answer {
		return true
	}
	// Keep fuzzy matching narrow: it only repairs one typo (including a
	// transposition) in a substantial word. This makes oral recall forgiving
	// without treating a different short answer as correct.
	return len(input) >= 5 && len(answer) >= 5 && damerauLevenshteinAtMostOne(input, answer)
}

func damerauLevenshteinAtMostOne(a, b string) bool {
	ar, br := []rune(a), []rune(b)
	if abs(len(ar)-len(br)) > 1 {
		return false
	}
	if len(ar) == len(br) {
		differences := make([]int, 0, 2)
		for i := range ar {
			if ar[i] != br[i] {
				differences = append(differences, i)
			}
		}
		if len(differences) <= 1 {
			return true
		}
		return len(differences) == 2 && differences[1] == differences[0]+1 && ar[differences[0]] == br[differences[1]] && ar[differences[1]] == br[differences[0]]
	}
	if len(ar) > len(br) {
		ar, br = br, ar
	}
	i, j, edits := 0, 0, 0
	for i < len(ar) && j < len(br) {
		if ar[i] == br[j] {
			i++
		} else {
			edits++
			j++
		}
		if edits > 1 {
			return false
		}
	}
	return true
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Normalize makes punctuation, common U.S. forms, number spellings, and
// leading articles irrelevant while keeping the actual answer words intact.
func Normalize(s string) string {
	var words []string
	var token strings.Builder
	flush := func() {
		if token.Len() == 0 {
			return
		}
		word := token.String()
		token.Reset()
		if n, err := strconv.Atoi(word); err == nil {
			words = append(words, numberWords(n)...)
			return
		}
		words = append(words, word)
	}
	for _, r := range strings.ToLower(s) {
		r = latinBase(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			token.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	for i := 0; i+1 < len(words); {
		if words[i] == "u" && words[i+1] == "s" {
			words = append(words[:i], append([]string{"united", "states"}, words[i+2:]...)...)
			i += 2
			continue
		}
		i++
	}
	for i, word := range words {
		// “Judiciary” is the ordinary noun for the judicial branch. This is a
		// matching alias only; it does not alter the USCIS answer text shown in
		// the study material.
		if word == "judiciary" {
			words[i] = "judicial"
		}
	}
	for len(words) > 0 && (words[0] == "a" || words[0] == "an" || words[0] == "the") {
		words = words[1:]
	}
	return strings.Join(words, " ")
}

func numberWords(n int) []string {
	if n < 0 || n > 999 {
		return []string{strconv.Itoa(n)}
	}
	ones := []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}
	tens := []string{"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}
	if n < 20 {
		return []string{ones[n]}
	}
	if n < 100 {
		if n%10 == 0 {
			return []string{tens[n/10]}
		}
		return []string{tens[n/10], ones[n%10]}
	}
	result := []string{ones[n/100], "hundred"}
	if n%100 != 0 {
		result = append(result, numberWords(n%100)...)
	}
	return result
}

func latinBase(r rune) rune {
	const accented = "àáâãäåāăąçćčďèéêëēĕėęěìíîïīĭįłñńňòóôõöøōŏőŕřśšşťùúûüūŭůűųýÿźž"
	const plain = "aaaaaaaaacccdeeeeeeeeeiiiiiiilnnnoooooooooorrssstuuuuuuuuuyyzz"
	for i, accentedRune := range []rune(accented) {
		if accentedRune == r {
			return []rune(plain)[i]
		}
	}
	return r
}
