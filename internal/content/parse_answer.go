package content

import (
	"fmt"
	"strings"
)

func reflow(s string) string { return strings.Join(strings.Fields(s), " ") }

// ParseAnswer preserves the reflowed source text and separates reader guidance
// before interpreting optional parentheses. Unbalanced notation is an error.
func ParseAnswer(text string) (Answer, error) {
	a := Answer{Text: text}
	var body, guidance strings.Builder
	inGuidance := false
	for _, r := range text {
		switch r {
		case '[':
			if inGuidance {
				return a, fmt.Errorf("nested guidance in %q", text)
			}
			inGuidance = true
			guidance.WriteRune(' ')
		case ']':
			if !inGuidance {
				return a, fmt.Errorf("unmatched guidance in %q", text)
			}
			inGuidance = false
			body.WriteRune(' ')
		default:
			if inGuidance {
				guidance.WriteRune(r)
			} else {
				body.WriteRune(r)
			}
		}
	}
	if inGuidance {
		return a, fmt.Errorf("unclosed guidance in %q", text)
	}
	var core, full strings.Builder
	depth := 0
	for _, r := range body.String() {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return a, fmt.Errorf("unmatched optional text in %q", text)
			}
		default:
			full.WriteRune(r)
			if depth == 0 {
				core.WriteRune(r)
			}
		}
	}
	if depth != 0 {
		return a, fmt.Errorf("unclosed optional text in %q", text)
	}
	a.Core, a.Full, a.Guidance = reflow(core.String()), reflow(full.String()), reflow(guidance.String())
	return a, nil
}
