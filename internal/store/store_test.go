package store

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestFileStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "n400", "progress.json")
	store := NewFile(path)
	want := Empty()
	want.Profile.State = "CA"
	want.Profile.ZIP = "94110"
	want.Profile.Overrides[38] = "Example President"
	want.Questions[1] = QuestionProgress{Correct: 2, Incorrect: 1}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Profile.State != want.Profile.State || got.Profile.ZIP != want.Profile.ZIP || got.Profile.Overrides[38] != want.Profile.Overrides[38] || got.Questions[1] != want.Questions[1] {
		t.Fatalf("round trip = %#v", got)
	}
}

func TestFileStoreCorruptFileRecoversEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := NewFile(path)
	got, err := store.Load()
	if err == nil {
		t.Fatal("corrupt file should be reported")
	}
	if len(got.Cards) != 0 || len(got.Questions) != 0 || len(got.Profile.Overrides) != 0 {
		t.Fatalf("corrupt recovery = %#v", got)
	}
	if err := store.Update(func(data *Data) { data.Profile.State = "CA" }); err != nil {
		t.Fatal(err)
	}
	got, err = store.Load()
	if err != nil || got.Profile.State != "CA" {
		t.Fatalf("recovered update = %#v, %v", got, err)
	}
}

func TestFileStoreFailedReplacePreservesPreviousFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	store := NewFile(path)
	if err := store.Update(func(data *Data) { data.Profile.State = "CA" }); err != nil {
		t.Fatal(err)
	}
	store.rename = func(string, string) error { return errors.New("disk full") }
	if err := store.Update(func(data *Data) { data.Profile.State = "NY" }); err == nil {
		t.Fatal("replace failure should be returned")
	}
	store.rename = os.Rename
	got, err := store.Load()
	if err != nil || got.Profile.State != "CA" {
		t.Fatalf("previous data should survive: %#v, %v", got, err)
	}
}

func TestFileStoreConcurrentUpdates(t *testing.T) {
	store := NewFile(filepath.Join(t.TempDir(), "progress.json"))
	const updates = 32
	var group sync.WaitGroup
	for i := 0; i < updates; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if err := store.Update(func(data *Data) {
				progress := data.Questions[1]
				progress.Correct++
				data.Questions[1] = progress
			}); err != nil {
				t.Errorf("update: %v", err)
			}
		}()
	}
	group.Wait()
	got, err := store.Load()
	if err != nil || got.Questions[1].Correct != updates {
		t.Fatalf("concurrent updates = %#v, %v", got, err)
	}
}
