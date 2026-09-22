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
