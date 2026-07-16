package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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

func TestStripClientPowerOn(t *testing.T) {
	method, path, _ := callStrip(t, func(c *StripClient) error {
		return c.On()
	})
	if method != http.MethodPost || path != "/power/on" {
		t.Fatalf("request = %s %s, want POST /power/on", method, path)
	}
}

func TestStripClientSetColor(t *testing.T) {
	_, path, body := callStrip(t, func(c *StripClient) error {
		return c.SetColor(StripOrange)
	})
	if path != "/color" {
		t.Fatalf("path = %s, want /color", path)
	}
	want := map[string]float64{"r": 255, "g": 128, "b": 0}
	for key, value := range want {
		if body[key] != value {
			t.Fatalf("body[%s] = %v, want %v", key, body[key], value)
		}
	}
}

func TestStripClientSetBrightness(t *testing.T) {
	_, path, body := callStrip(t, func(c *StripClient) error {
		return c.SetBrightness(75)
	})
	if path != "/brightness" {
		t.Fatalf("path = %s, want /brightness", path)
	}
	if body["value"] != float64(75) {
		t.Fatalf("value = %v, want 75", body["value"])
	}
}

func TestStripClientSetTimer(t *testing.T) {
	_, path, body := callStrip(t, func(c *StripClient) error {
		return c.SetTimer(30 * time.Minute)
	})
	if path != "/timer" {
		t.Fatalf("path = %s, want /timer", path)
	}
	if body["seconds"] != float64(1800) {
		t.Fatalf("seconds = %v, want 1800", body["seconds"])
	}
}

func TestStripClientCancelTimer(t *testing.T) {
	method, path, _ := callStrip(t, func(c *StripClient) error {
		return c.CancelTimer()
	})
	if method != http.MethodDelete || path != "/timer" {
		t.Fatalf("request = %s %s, want DELETE /timer", method, path)
	}
}

func TestStripClientStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/status" {
			t.Fatalf("request = %s %s, want GET /status", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(StripStatus{Connected: true, Powered: true})
	}))
	defer server.Close()

	status, err := newStripClient(server.URL).Status()
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.Connected || !status.Powered {
		t.Fatalf("status = %#v, want connected and powered", status)
	}
}

func TestStripClientRejectsInvalidValues(t *testing.T) {
	client := newStripClient("http://127.0.0.1")
	if err := client.SetRGB(256, 0, 0); err == nil {
		t.Fatal("expected RGB error")
	}
	if err := client.SetBrightness(101); err == nil {
		t.Fatal("expected brightness error")
	}
	if err := client.SetMode(213); err == nil {
		t.Fatal("expected mode error")
	}
}

func callStrip(t *testing.T, call func(*StripClient) error) (string, string, map[string]float64) {
	t.Helper()
	var method, path string
	body := map[string]float64{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := call(newStripClient(server.URL)); err != nil {
		t.Fatalf("strip call error = %v", err)
	}
	return method, path, body
}
