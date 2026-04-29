package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Note struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type NotesStore struct {
	path  string
	mu    sync.Mutex
	notes []Note
}

func newNotesStore(path string) (*NotesStore, error) {
	s := &NotesStore{path: path, notes: []Note{}}
	if b, err := os.ReadFile(path); err == nil && len(b) > 0 {
		if err := json.Unmarshal(b, &s.notes); err != nil {
			return nil, err
		}
		// Migrate old notes without ID.
		changed := false
		for i := range s.notes {
			if strings.TrimSpace(s.notes[i].ID) == "" {
				s.notes[i].ID = fmt.Sprintf("note-%d", time.Now().UnixNano()+int64(i))
				changed = true
			}
		}
		if changed {
			if err := s.saveLocked(); err != nil {
				return nil, err
			}
		}
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *NotesStore) saveLocked() error {
	b, err := json.MarshalIndent(s.notes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o644)
}

func (s *NotesStore) Add(text string) (Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := Note{ID: fmt.Sprintf("note-%d", time.Now().UnixNano()), Text: text, CreatedAt: time.Now().UTC()}
	s.notes = append(s.notes, n)
	if err := s.saveLocked(); err != nil {
		return Note{}, err
	}
	return n, nil
}

func (s *NotesStore) DeleteByID(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id = strings.TrimSpace(id)
	if id == "" {
		return false, nil
	}
	for i, n := range s.notes {
		if n.ID == id {
			s.notes = append(s.notes[:i], s.notes[i+1:]...)
			if err := s.saveLocked(); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func (s *NotesStore) Last(limit int) []Note {
	s.mu.Lock()
	defer s.mu.Unlock()
	if limit <= 0 || limit > len(s.notes) {
		limit = len(s.notes)
	}
	out := make([]Note, 0, limit)
	for i := len(s.notes) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, s.notes[i])
	}
	return out
}

func (s *NotesStore) All() []Note {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Note, 0, len(s.notes))
	out = append(out, s.notes...)
	return out
}
