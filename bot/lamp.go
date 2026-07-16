package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type LampState struct {
	On bool `json:"on"`
}

type LampStore struct {
	path string
	mu   sync.Mutex
}

func newLampStore(path string) *LampStore {
	return &LampStore{path: path}
}

func (s *LampStore) IsOn() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if len(b) == 0 {
		return false, nil
	}
	var state LampState
	if err := json.Unmarshal(b, &state); err != nil {
		return false, err
	}
	return state.On, nil
}

func (s *LampStore) Set(on bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	b, err := json.MarshalIndent(LampState{On: on}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o644)
}

type commandRunner func(name string, args ...string) (string, error)

type LampService struct {
	location string
	port     string
	store    *LampStore
	run      commandRunner
	mu       sync.Mutex
}

func newLampService(location, port string, store *LampStore) *LampService {
	return &LampService{
		location: location,
		port:     port,
		store:    store,
		run:      runCommand,
	}
}

func runCommand(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (s *LampService) Toggle() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	on, err := s.store.IsOn()
	if err != nil {
		return false, fmt.Errorf("read lamp state: %w", err)
	}

	next := !on
	action := "off"
	if next {
		action = "on"
	}

	out, err := s.run("uhubctl", "-l", s.location, "-p", s.port, "-a", action)
	if err != nil {
		if out == "" {
			return false, fmt.Errorf("switch lamp %s: %w", action, err)
		}
		return false, fmt.Errorf("switch lamp %s: %w: %s", action, err, out)
	}
	if err := s.store.Set(next); err != nil {
		return false, fmt.Errorf("save lamp state: %w", err)
	}
	return next, nil
}
