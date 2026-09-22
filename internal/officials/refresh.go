package officials

import (
	"context"
	"fmt"
	"time"
)

// Refresh fetches the four federal offices and writes them to the local
// sidecar overlay. It runs only when a caller explicitly invokes it — never
// on application startup or a study page — and its failure is reported
// rather than silently swallowed; the previous sidecar value, if any, is left
// untouched on failure.
func Refresh(ctx context.Context, client FederalClient, sidecarPath string, now time.Time) (Sidecar, error) {
	answers, err := client.Fetch(ctx)
	if err != nil {
		return Sidecar{}, fmt.Errorf("refreshing federal officials: %w", err)
	}
	sidecar := Sidecar{AsOf: now.Format("2006-01-02"), Source: "Wikidata", Federal: answers}
	if err := SaveSidecar(sidecarPath, sidecar); err != nil {
		return Sidecar{}, err
	}
	return sidecar, nil
}
