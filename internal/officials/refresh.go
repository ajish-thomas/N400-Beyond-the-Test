package officials

import (
	"context"
	"fmt"
	"time"
)

// Refresh fetches the four federal offices, plus the governor of state (when
// state has a Wikidata identifier — D.C. deliberately does not, since it has
// no governor), and writes them to the local sidecar overlay. It runs only
// when a caller explicitly invokes it — never on application startup or a
// study page — and its failure is reported rather than silently swallowed;
// the previous sidecar value, if any, is left untouched on failure. A
// previously refreshed governor for a different state is preserved.
func Refresh(ctx context.Context, federal FederalClient, governor GovernorClient, state State, sidecarPath string, now time.Time) (Sidecar, error) {
	existing, err := LoadSidecar(sidecarPath)
	if err != nil {
		return Sidecar{}, err
	}
	answers, err := federal.Fetch(ctx)
	if err != nil {
		return Sidecar{}, fmt.Errorf("refreshing federal officials: %w", err)
	}
	sidecar := Sidecar{AsOf: now.Format("2006-01-02"), Source: "Wikidata", Federal: answers, Governors: existing.Governors}
	if state.QID != "" {
		name, err := governor.Fetch(ctx, state.QID)
		if err != nil {
			return Sidecar{}, fmt.Errorf("refreshing governor: %w", err)
		}
		if sidecar.Governors == nil {
			sidecar.Governors = make(map[string]string)
		}
		sidecar.Governors[state.Code] = name
	}
	if err := SaveSidecar(sidecarPath, sidecar); err != nil {
		return Sidecar{}, err
	}
	return sidecar, nil
}
