package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestLampStoreDefaultOff(t *testing.T) {
	store := newLampStore(t.TempDir() + "/lamp_state.json")

	on, err := store.IsOn()
	if err != nil {
		t.Fatalf("IsOn() error = %v", err)
	}
	if on {
		t.Fatal("expected missing state file to mean off")
	}
}

func TestLampStoreSaveAndRead(t *testing.T) {
	store := newLampStore(t.TempDir() + "/nested/lamp_state.json")

	if err := store.Set(true); err != nil {
		t.Fatalf("Set(true) error = %v", err)
	}
	on, err := store.IsOn()
	if err != nil {
		t.Fatalf("IsOn() error = %v", err)
	}
	if !on {
		t.Fatal("expected saved state to be on")
	}
}

func TestLampServiceToggleOffToOn(t *testing.T) {
	store := newLampStore(t.TempDir() + "/lamp_state.json")
	service := newLampService("1-1", "4", store)

	var gotName string
	var gotArgs []string
	service.run = func(name string, args ...string) error {
		gotName = name
		gotArgs = append([]string{}, args...)
		return nil
	}

	on, err := service.Toggle()
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}
	if !on {
		t.Fatal("expected lamp to be on after first toggle")
	}
	if gotName != "uhubctl" {
		t.Fatalf("command name = %q, want uhubctl", gotName)
	}
	wantArgs := []string{"-l", "1-1", "-p", "4", "-a", "on"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}

func TestLampServiceToggleOnToOff(t *testing.T) {
	store := newLampStore(t.TempDir() + "/lamp_state.json")
	if err := store.Set(true); err != nil {
		t.Fatalf("Set(true) error = %v", err)
	}
	service := newLampService("1-1", "4", store)

	var gotArgs []string
	service.run = func(name string, args ...string) error {
		gotArgs = append([]string{}, args...)
		return nil
	}

	on, err := service.Toggle()
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}
	if on {
		t.Fatal("expected lamp to be off after toggle")
	}
	wantArgs := []string{"-l", "1-1", "-p", "4", "-a", "off"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}

func TestLampServiceDoesNotSaveOnCommandError(t *testing.T) {
	store := newLampStore(t.TempDir() + "/lamp_state.json")
	service := newLampService("1-1", "4", store)
	service.run = func(name string, args ...string) error {
		return errors.New("failed")
	}

	if _, err := service.Toggle(); err == nil {
		t.Fatal("expected toggle error")
	}
	on, err := store.IsOn()
	if err != nil {
		t.Fatalf("IsOn() error = %v", err)
	}
	if on {
		t.Fatal("state should remain off after failed command")
	}
}
