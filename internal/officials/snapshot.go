// Package officials resolves the eight civics answers that change over time or
// with a learner's state. It never replaces the USCIS verification notice.
package officials

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed data/officials.json data/states.json
var data embed.FS

type Snapshot struct {
	AsOf    string            `json:"as_of"`
	Federal map[string]string `json:"federal"`
	States  []State           `json:"states"`
	Sources map[string]string `json:"sources"`
}

type State struct {
	Code     string   `json:"code"`
	Name     string   `json:"name"`
	Capital  string   `json:"capital"`
	Senators []string `json:"senators"`
}

// LoadSnapshot reads the dated offline values shipped in the binary.
func LoadSnapshot() (Snapshot, error) {
	b, err := data.ReadFile("data/officials.json")
	if err != nil {
		return Snapshot{}, err
	}
	var snapshot Snapshot
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("decoding officials snapshot: %w", err)
	}
	if snapshot.AsOf == "" || len(snapshot.Federal) != 4 || len(snapshot.Sources) != 4 {
		return Snapshot{}, fmt.Errorf("officials snapshot is incomplete")
	}
	b, err = data.ReadFile("data/states.json")
	if err != nil {
		return Snapshot{}, err
	}
	var states []State
	decoder = json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&states); err != nil {
		return Snapshot{}, fmt.Errorf("decoding states snapshot: %w", err)
	}
	seen := map[string]bool{}
	for i := range states {
		state := &states[i]
		state.Senators = currentSenators[state.Code]
		if len(state.Code) != 2 || state.Name == "" || state.Capital == "" || seen[state.Code] || (state.Code != "DC" && len(state.Senators) != 2) || (state.Code == "DC" && len(state.Senators) != 0) {
			return Snapshot{}, fmt.Errorf("invalid or duplicate state %q", state.Code)
		}
		seen[state.Code] = true
	}
	if len(states) != 51 || !seen["DC"] {
		return Snapshot{}, fmt.Errorf("states snapshot must contain 50 states and DC")
	}
	snapshot.States = states
	return snapshot, nil
}

type Answer struct {
	Text      string
	Source    string
	SourceURL string
	AsOf      string
	Available bool
}

// Resolver applies the required precedence: a manual entry, then a local live
// sidecar, then the dated bundled snapshot. Sidecar loading belongs to the
// refresh client so network data never reaches a request path.
type Resolver struct {
	Snapshot Snapshot
	Manual   map[int]string
	Sidecar  map[int]string
}

func (r Resolver) Resolve(questionID int, stateCode string) Answer {
	answers := r.ResolveAll(questionID, stateCode)
	if len(answers) == 0 {
		return Answer{}
	}
	return answers[0]
}

// ResolveAll returns every equally acceptable resolved response. Q23 accepts
// either of a state's two senators, so reducing it to one value would grade a
// valid response incorrectly.
func (r Resolver) ResolveAll(questionID int, stateCode string) []Answer {
	if answer := strings.TrimSpace(r.Manual[questionID]); answer != "" {
		return []Answer{{Text: answer, Source: "manual entry", Available: true}}
	}
	if answer := strings.TrimSpace(r.Sidecar[questionID]); answer != "" {
		return []Answer{{Text: answer, Source: "local refreshed data", Available: true}}
	}
	if questionID == 23 || questionID == 62 {
		for _, state := range r.Snapshot.States {
			if state.Code == strings.ToUpper(strings.TrimSpace(stateCode)) {
				if questionID == 23 {
					answers := make([]Answer, 0, len(state.Senators))
					for _, senator := range state.Senators {
						answers = append(answers, Answer{Text: senator, Source: "bundled Senate roster", SourceURL: "https://www.senate.gov/senators/index.htm?State=", AsOf: r.Snapshot.AsOf, Available: true})
					}
					return answers
				}
				return []Answer{{Text: state.Capital, Source: "bundled state snapshot", AsOf: r.Snapshot.AsOf, Available: true}}
			}
		}
		return nil
	}
	key := map[int]string{30: "speaker", 38: "president", 39: "vice_president", 57: "chief_justice"}[questionID]
	if answer := strings.TrimSpace(r.Snapshot.Federal[key]); answer != "" {
		return []Answer{{Text: answer, Source: "bundled officials snapshot", SourceURL: r.Snapshot.Sources[key], AsOf: r.Snapshot.AsOf, Available: true}}
	}
	return nil
}
