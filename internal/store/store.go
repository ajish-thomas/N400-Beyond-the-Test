// Package store persists private study progress on the local machine.
package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"n400/internal/flashcard"
)

// Data is the complete locally stored study record. It deliberately contains
// no street address: district lookups may use one transiently but never save it.
type Data struct {
	Cards     map[int]flashcard.Card   `json:"cards"`
	Questions map[int]QuestionProgress `json:"questions"`
	Profile   Profile                  `json:"profile"`
}

type QuestionProgress struct {
	Correct      int       `json:"correct"`
	Incorrect    int       `json:"incorrect"`
	LastAnswered time.Time `json:"last_answered"`
}

type Profile struct {
	State     string         `json:"state"`
	ZIP       string         `json:"zip"`
	District  string         `json:"district"`
	Overrides map[int]string `json:"overrides"`
}

func Empty() Data {
	return Data{
		Cards:     make(map[int]flashcard.Card),
		Questions: make(map[int]QuestionProgress),
		Profile:   Profile{Overrides: make(map[int]string)},
	}
}

func (d *Data) normalize() {
	if d.Cards == nil {
		d.Cards = make(map[int]flashcard.Card)
	}
	if d.Questions == nil {
		d.Questions = make(map[int]QuestionProgress)
	}
	if d.Profile.Overrides == nil {
		d.Profile.Overrides = make(map[int]string)
	}
}

// FileStore reads and atomically writes one JSON progress file.
type FileStore struct {
	path   string
	mu     sync.Mutex
	rename func(string, string) error
}

func NewFile(path string) *FileStore {
	return &FileStore{path: path, rename: os.Rename}
}

// Load returns empty progress for a missing file. A corrupt file also returns
// empty progress plus an error, allowing the caller to keep the app usable and
// explain that saved data could not be read.
func (s *FileStore) Load() (Data, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *FileStore) loadLocked() (Data, error) {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return Empty(), nil
	}
	if err != nil {
		return Empty(), fmt.Errorf("reading progress: %w", err)
	}
	data := Empty()
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		return Empty(), fmt.Errorf("decoding progress: %w", err)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return Empty(), fmt.Errorf("decoding progress: trailing data")
	}
	data.normalize()
	return data, nil
}

// Save replaces the complete record atomically. On a failure before rename,
// the prior valid file remains intact.
func (s *FileStore) Save(data Data) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(data)
}

func (s *FileStore) saveLocked(data Data) error {
	data.normalize()
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating progress directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".progress-*")
	if err != nil {
		return fmt.Errorf("creating progress temporary file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	encoder := json.NewEncoder(tmp)
	if err := encoder.Encode(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing progress: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing progress: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing progress: %w", err)
	}
	if err := s.rename(tmpName, s.path); err != nil {
		return fmt.Errorf("replacing progress atomically: %w", err)
	}
	return nil
}

// Update reads, changes, and persists progress while holding one mutex, so two
// simultaneous study actions cannot overwrite one another's changes.
func (s *FileStore) Update(change func(*Data)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := s.loadLocked()
	if err != nil {
		// A corrupt record must not make the application unusable. The next
		// successful write replaces it with a valid, empty record.
		data = Empty()
	}
	change(&data)
	return s.saveLocked(data)
}
