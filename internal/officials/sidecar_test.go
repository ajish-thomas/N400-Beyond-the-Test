package officials

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSidecarRoundTripAndReset(t *testing.T) {
	path := filepath.Join(t.TempDir(), "n400", "officials-live.json")
	if sidecar, err := LoadSidecar(path); err != nil || sidecar.AsOf != "" {
		t.Fatalf("missing sidecar should load empty: %+v %v", sidecar, err)
	}
	want := Sidecar{AsOf: "2026-09-22", Source: "Wikidata", Federal: map[string]string{"president": "Live President"}}
	if err := SaveSidecar(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSidecar(path)
	if err != nil || got.AsOf != want.AsOf || got.Source != want.Source || got.Federal["president"] != "Live President" {
		t.Fatalf("round trip = %+v, err %v", got, err)
	}
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Fatalf("SaveSidecar did not create its directory: %v", err)
	}
	if err := ResetSidecar(path); err != nil {
		t.Fatal(err)
	}
	if sidecar, err := LoadSidecar(path); err != nil || sidecar.AsOf != "" {
		t.Fatalf("reset sidecar should load empty: %+v %v", sidecar, err)
	}
	if err := ResetSidecar(path); err != nil {
		t.Fatalf("resetting an already-absent sidecar must not error: %v", err)
	}
}

func TestSidecarRejectsCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "officials-live.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSidecar(path); err == nil {
		t.Fatal("expected a decode error for corrupt sidecar")
	}
}

func TestSidecarOverridesMapsOfficeKeysToQuestionIDs(t *testing.T) {
	sidecar := Sidecar{Federal: map[string]string{"president": "Live President", "vice_president": "Live VP"}}
	overrides := sidecar.Overrides("")
	if overrides[38] != "Live President" || overrides[39] != "Live VP" || len(overrides) != 2 {
		t.Fatalf("overrides = %#v", overrides)
	}
}

func TestSidecarOverridesIncludesGovernorForSelectedStateOnly(t *testing.T) {
	sidecar := Sidecar{Governors: map[string]string{"CA": "Live California Governor", "TX": "Live Texas Governor"}}
	if overrides := sidecar.Overrides("CA"); overrides[61] != "Live California Governor" || len(overrides) != 1 {
		t.Fatalf("CA overrides = %#v", overrides)
	}
	if overrides := sidecar.Overrides("tx"); overrides[61] != "Live Texas Governor" {
		t.Fatalf("lowercase state code should still match: %#v", overrides)
	}
	if overrides := sidecar.Overrides("NY"); overrides[61] != "" {
		t.Fatalf("a state with no refreshed governor must not resolve: %#v", overrides)
	}
}
