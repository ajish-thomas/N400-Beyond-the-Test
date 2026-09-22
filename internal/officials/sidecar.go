package officials

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Sidecar is the local overlay for the four federal offices that change over
// time. A refresh writes it; it never overwrites the embedded bundled
// snapshot, and its absence is not an error — the app resolves from the
// bundled snapshot alone until a refresh has ever run.
type Sidecar struct {
	AsOf    string            `json:"as_of"`
	Source  string            `json:"source"`
	Federal map[string]string `json:"federal"`
}

// SidecarPath returns the local refresh overlay's location under a user
// config directory (typically os.UserConfigDir()).
func SidecarPath(configDir string) string {
	return filepath.Join(configDir, "n400", "officials-live.json")
}

// LoadSidecar reads the local refresh overlay. A missing file returns a zero
// Sidecar and no error.
func LoadSidecar(path string) (Sidecar, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Sidecar{}, nil
	}
	if err != nil {
		return Sidecar{}, fmt.Errorf("reading local refreshed officials: %w", err)
	}
	var sidecar Sidecar
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&sidecar); err != nil {
		return Sidecar{}, fmt.Errorf("decoding local refreshed officials: %w", err)
	}
	return sidecar, nil
}

// SaveSidecar atomically writes a freshly fetched overlay.
func SaveSidecar(path string, sidecar Sidecar) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating local settings directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".officials-live-*")
	if err != nil {
		return fmt.Errorf("creating local refreshed officials file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	encoder := json.NewEncoder(tmp)
	if err := encoder.Encode(sidecar); err != nil {
		tmp.Close()
		return fmt.Errorf("writing local refreshed officials: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing local refreshed officials: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing local refreshed officials: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replacing local refreshed officials atomically: %w", err)
	}
	return nil
}

// ResetSidecar removes the local overlay, so resolution falls back to the
// bundled snapshot. Removing an already-absent file is not an error.
func ResetSidecar(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing local refreshed officials: %w", err)
	}
	return nil
}

// Overrides converts the sidecar's federal answers into the question-ID-keyed
// map Resolver.Sidecar expects.
func (s Sidecar) Overrides() map[int]string {
	overrides := make(map[int]string, len(federalQuestionKeys))
	for questionID, key := range federalQuestionKeys {
		if answer := s.Federal[key]; answer != "" {
			overrides[questionID] = answer
		}
	}
	return overrides
}
