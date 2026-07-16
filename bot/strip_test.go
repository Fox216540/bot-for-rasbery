package main

import (
	"errors"
	"testing"
	"time"
)

type fakeStripLamp struct {
	onCalls        []bool
	rgbCalls       [][3]int
	brightnessCall []int
	speedCalls     []int
	modeCalls      []int
	closeCalls     int
	err            error
}

type fakeStripConnector struct {
	lamp  stripLamp
	errs  []error
	calls int
}

func (f *fakeStripConnector) Connect(addr, name string) (stripLamp, error) {
	f.calls++
	if len(f.errs) >= f.calls && f.errs[f.calls-1] != nil {
		return nil, f.errs[f.calls-1]
	}
	return f.lamp, nil
}

func (f *fakeStripLamp) LightOn(on bool) error {
	f.onCalls = append(f.onCalls, on)
	return f.err
}

func (f *fakeStripLamp) ChangeColorRGB(r, g, b int) error {
	f.rgbCalls = append(f.rgbCalls, [3]int{r, g, b})
	return f.err
}

func (f *fakeStripLamp) ChangeBrightness(brightness, lightMode int) error {
	f.brightnessCall = append(f.brightnessCall, brightness)
	return f.err
}

func (f *fakeStripLamp) ChangeMode(mode int) error {
	f.modeCalls = append(f.modeCalls, mode)
	return f.err
}

func (f *fakeStripLamp) ChangeModeSpeed(speed int) error {
	f.speedCalls = append(f.speedCalls, speed)
	return f.err
}

func (f *fakeStripLamp) Close() error {
	f.closeCalls++
	return nil
}

func TestParseRGB(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantR     int
		wantG     int
		wantB     int
		wantError bool
	}{
		{name: "space separated", input: "255 128 0", wantR: 255, wantG: 128, wantB: 0},
		{name: "hex", input: "#FF8000", wantR: 255, wantG: 128, wantB: 0},
		{name: "lower hex", input: "#ff8000", wantR: 255, wantG: 128, wantB: 0},
		{name: "short hex", input: "#F80", wantError: true},
		{name: "out of range", input: "256 0 0", wantError: true},
		{name: "missing part", input: "255 0", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, gotG, gotB, err := parseRGB(tt.input)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseRGB() error = %v", err)
			}
			if gotR != tt.wantR || gotG != tt.wantG || gotB != tt.wantB {
				t.Fatalf("parseRGB() = %d %d %d, want %d %d %d", gotR, gotG, gotB, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestBrightnessValue(t *testing.T) {
	tests := map[int]int{
		0:   0,
		10:  25,
		50:  127,
		100: 255,
	}
	for input, want := range tests {
		got, err := brightnessValue(input)
		if err != nil {
			t.Fatalf("brightnessValue(%d) error = %v", input, err)
		}
		if got != want {
			t.Fatalf("brightnessValue(%d) = %d, want %d", input, got, want)
		}
	}
	if _, err := brightnessValue(101); err == nil {
		t.Fatal("expected error")
	}
}

func TestStripServiceReusesConnection(t *testing.T) {
	lamp := &fakeStripLamp{}
	service := newStripService("AA:BB")
	connector := &fakeStripConnector{lamp: lamp}
	service.connector = connector

	if err := service.On(); err != nil {
		t.Fatalf("On() error = %v", err)
	}
	if err := service.Off(); err != nil {
		t.Fatalf("Off() error = %v", err)
	}
	if len(lamp.onCalls) != 2 || !lamp.onCalls[0] || lamp.onCalls[1] {
		t.Fatalf("onCalls = %#v, want true then false", lamp.onCalls)
	}
	if connector.calls != 1 {
		t.Fatalf("connect calls = %d, want 1", connector.calls)
	}
	if lamp.closeCalls != 0 {
		t.Fatalf("closeCalls = %d, want 0", lamp.closeCalls)
	}
}

func TestStripServiceRetriesConnect(t *testing.T) {
	lamp := &fakeStripLamp{}
	service := newStripService("AA:BB")
	connector := &fakeStripConnector{
		lamp: lamp,
		errs: []error{errors.New("connect failed")},
	}
	service.connector = connector

	if err := service.Off(); err != nil {
		t.Fatalf("Off() error = %v", err)
	}
	if connector.calls != 2 {
		t.Fatalf("connect calls = %d, want 2", connector.calls)
	}
	if lamp.closeCalls != 0 {
		t.Fatalf("closeCalls = %d, want 0", lamp.closeCalls)
	}
}

func TestStripServiceReconnectsOnCommandError(t *testing.T) {
	first := &fakeStripLamp{err: errors.New("write failed")}
	second := &fakeStripLamp{}
	service := newStripService("AA:BB")
	connector := &sequenceStripConnector{lamps: []stripLamp{first, second}}
	service.connector = connector

	if err := service.On(); err != nil {
		t.Fatalf("On() error = %v", err)
	}
	if connector.calls != 2 {
		t.Fatalf("connect calls = %d, want 2", connector.calls)
	}
	if first.closeCalls != 1 {
		t.Fatalf("first closeCalls = %d, want 1", first.closeCalls)
	}
	if len(second.onCalls) != 1 || !second.onCalls[0] {
		t.Fatalf("second onCalls = %#v, want true call", second.onCalls)
	}
}

func TestStripServiceRejectsInvalidMode(t *testing.T) {
	service := newStripService("AA:BB")
	if err := service.SetMode(0); err == nil {
		t.Fatal("expected error for mode 0")
	}
	if err := service.SetMode(128); err == nil {
		t.Fatal("expected error for mode 128")
	}
}

func TestStripServiceTimerReplacementAndCancel(t *testing.T) {
	lamp := &fakeStripLamp{}
	service := newStripService("AA:BB")
	service.connector = &fakeStripConnector{lamp: lamp}

	if err := service.SetTimer(time.Hour); err != nil {
		t.Fatalf("SetTimer() error = %v", err)
	}
	first := service.timer
	if err := service.SetTimer(time.Hour); err != nil {
		t.Fatalf("SetTimer() replacement error = %v", err)
	}
	if service.timer == nil {
		t.Fatal("expected active timer")
	}
	if service.timer == first {
		t.Fatal("expected timer replacement")
	}
	if err := service.CancelTimer(); err != nil {
		t.Fatalf("CancelTimer() error = %v", err)
	}
	if service.timer != nil {
		t.Fatal("expected nil timer after cancel")
	}
}

func TestStripServiceCloseCancelsTimerAndDisconnects(t *testing.T) {
	lamp := &fakeStripLamp{}
	service := newStripService("AA:BB")
	service.connector = &fakeStripConnector{lamp: lamp}

	if err := service.On(); err != nil {
		t.Fatalf("On() error = %v", err)
	}
	if err := service.SetTimer(time.Hour); err != nil {
		t.Fatalf("SetTimer() error = %v", err)
	}
	if err := service.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if service.timer != nil {
		t.Fatal("expected nil timer after close")
	}
	if lamp.closeCalls != 1 {
		t.Fatalf("closeCalls = %d, want 1", lamp.closeCalls)
	}
}

func TestStripServiceIdleTimeoutDisconnects(t *testing.T) {
	lamp := &fakeStripLamp{}
	service := newStripService("AA:BB")
	service.connector = &fakeStripConnector{lamp: lamp}
	service.idleTTL = time.Millisecond

	if err := service.On(); err != nil {
		t.Fatalf("On() error = %v", err)
	}

	deadline := time.After(200 * time.Millisecond)
	for lamp.closeCalls == 0 {
		select {
		case <-deadline:
			t.Fatal("expected idle disconnect")
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

func TestStripServiceCommandResetsIdleTimeout(t *testing.T) {
	lamp := &fakeStripLamp{}
	service := newStripService("AA:BB")
	connector := &fakeStripConnector{lamp: lamp}
	service.connector = connector
	service.idleTTL = 50 * time.Millisecond

	if err := service.On(); err != nil {
		t.Fatalf("On() error = %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := service.Off(); err != nil {
		t.Fatalf("Off() error = %v", err)
	}
	if connector.calls != 1 {
		t.Fatalf("connect calls = %d, want 1", connector.calls)
	}
	if lamp.closeCalls != 0 {
		t.Fatalf("closeCalls before idle = %d, want 0", lamp.closeCalls)
	}
}

type sequenceStripConnector struct {
	lamps []stripLamp
	calls int
}

func (s *sequenceStripConnector) Connect(addr, name string) (stripLamp, error) {
	s.calls++
	if s.calls > len(s.lamps) {
		return nil, errors.New("no lamp")
	}
	return s.lamps[s.calls-1], nil
}
