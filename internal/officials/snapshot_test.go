package officials

import "testing"

func TestSnapshotAndFederalResolution(t *testing.T) {
	snapshot, err := LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	resolver := Resolver{Snapshot: snapshot}
	for _, id := range []int{30, 38, 39, 57} {
		answer := resolver.Resolve(id, "")
		if !answer.Available || answer.Text == "" || answer.AsOf != snapshot.AsOf {
			t.Fatalf("question %d resolution = %#v", id, answer)
		}
	}
	if answer := resolver.Resolve(62, "CA"); !answer.Available || answer.Text != "Sacramento" {
		t.Fatalf("California capital = %#v", answer)
	}
	if senators := resolver.ResolveAll(23, "CA"); len(senators) != 2 || senators[0].Text != "Alex Padilla" || senators[1].Text != "Adam B. Schiff" {
		t.Fatalf("California senators = %#v", senators)
	}
	if answer := resolver.Resolve(62, "ZZ"); answer.Available {
		t.Fatalf("unknown state resolved = %#v", answer)
	}
}

func TestDistrictCandidatesPreserveMultiDistrictZIPs(t *testing.T) {
	snapshot, err := LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	candidates := snapshot.DistrictCandidates("90002")
	if len(candidates) != 4 || candidates[0] != "37" || candidates[3] != "44" {
		t.Fatalf("90002 candidates = %v", candidates)
	}
	if got := snapshot.DistrictCandidates("00000"); got != nil {
		t.Fatalf("unknown ZIP candidates = %v", got)
	}
}

func TestResolverPrecedence(t *testing.T) {
	snapshot, err := LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	resolver := Resolver{Snapshot: snapshot, Sidecar: map[int]string{38: "Sidecar President"}, SidecarAsOf: "2026-09-22", Manual: map[int]string{38: "Manual President"}}
	answer := resolver.Resolve(38, "")
	if answer.Text != "Manual President" || answer.Source != "manual entry" {
		t.Fatalf("manual precedence = %#v", answer)
	}
	delete(resolver.Manual, 38)
	answer = resolver.Resolve(38, "")
	if answer.Text != "Sidecar President" || answer.Source != "local refreshed data" || answer.AsOf != "2026-09-22" {
		t.Fatalf("sidecar precedence = %#v", answer)
	}
}

func TestRepresentativeResolvesFromBundledHouseRoster(t *testing.T) {
	snapshot, err := LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if name, ok := snapshot.Representative("CA", "37"); !ok || name != "Sydney Kamlager-Dove" {
		t.Fatalf("CA-37 representative = %q, %v", name, ok)
	}
	// Wyoming has a single at-large seat; the crosswalk and congress-legislators
	// agree on "00" for it, but the state-only fallback should still resolve it
	// even if an unrelated or blank district string is passed.
	if name, ok := snapshot.Representative("WY", ""); !ok || name != "Harriet M. Hageman" {
		t.Fatalf("Wyoming at-large representative = %q, %v", name, ok)
	}
	// D.C.'s crosswalk-derived district ("98") does not match
	// congress-legislators' own numbering for its delegate ("00"); the
	// single-seat fallback must resolve it anyway.
	if name, ok := snapshot.Representative("DC", "98"); !ok || name != "Eleanor Holmes Norton" {
		t.Fatalf("D.C. delegate = %q, %v", name, ok)
	}
	if _, ok := snapshot.Representative("CA", "99"); ok {
		t.Fatal("an unknown district in a multi-district state must not resolve")
	}
	if _, ok := snapshot.Representative("", "37"); ok {
		t.Fatal("an empty state must not resolve")
	}
}

func TestRepresentativeResolvesThroughResolver(t *testing.T) {
	snapshot, err := LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	resolver := Resolver{Snapshot: snapshot, District: "37"}
	if answer := resolver.Resolve(29, "CA"); !answer.Available || answer.Text != "Sydney Kamlager-Dove" || answer.Source != "bundled House roster" || answer.AsOf != snapshot.AsOf {
		t.Fatalf("California representative = %#v", answer)
	}
	noDistrict := Resolver{Snapshot: snapshot}
	if answer := noDistrict.Resolve(29, "CA"); answer.Available {
		t.Fatalf("a multi-district state with no district selected must be unavailable: %#v", answer)
	}
	if answer := noDistrict.Resolve(29, "WY"); !answer.Available || answer.Text != "Harriet M. Hageman" {
		t.Fatalf("an at-large state should resolve without a district selected: %#v", answer)
	}
	overridden := Resolver{Snapshot: snapshot, District: "37", Manual: map[int]string{29: "Manual Representative"}}
	if answer := overridden.Resolve(29, "CA"); answer.Text != "Manual Representative" || answer.Source != "manual entry" {
		t.Fatalf("manual override must still win: %#v", answer)
	}
}

func TestGovernorResolvesFromSidecarOverridesForSelectedStateOnly(t *testing.T) {
	snapshot, err := LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	sidecar := Sidecar{AsOf: "2026-09-22", Governors: map[string]string{"CA": "Live California Governor"}}
	resolver := Resolver{Snapshot: snapshot, Sidecar: sidecar.Overrides("CA"), SidecarAsOf: sidecar.AsOf}
	if answer := resolver.Resolve(61, "CA"); answer.Text != "Live California Governor" || answer.Source != "local refreshed data" || answer.AsOf != "2026-09-22" {
		t.Fatalf("California governor = %#v", answer)
	}
	unrefreshed := Resolver{Snapshot: snapshot, Sidecar: sidecar.Overrides("TX")}
	if answer := unrefreshed.Resolve(61, "TX"); answer.Available {
		t.Fatalf("a state with no refreshed governor must be unavailable: %#v", answer)
	}
	if answer := unrefreshed.Resolve(61, "DC"); answer.Available {
		t.Fatalf("D.C. has no governor and must be unavailable: %#v", answer)
	}
}
