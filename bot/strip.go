package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type RGB struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

var (
	StripRed    = RGB{255, 0, 0}
	StripGreen  = RGB{0, 255, 0}
	StripBlue   = RGB{0, 0, 255}
	StripYellow = RGB{255, 255, 0}
	StripPurple = RGB{128, 0, 255}
	StripWhite  = RGB{255, 255, 255}
	StripOrange = RGB{255, 128, 0}
)

type StripStatus struct {
	Connected bool `json:"connected"`
	Powered   bool `json:"powered"`
}

type StripClient struct {
	baseURL string
	client  *http.Client
}

func newStripClient(baseURL string) *StripClient {
	return &StripClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *StripClient) On() error {
	return c.post(context.Background(), "/power/on", nil)
}

func (c *StripClient) Off() error {
	return c.post(context.Background(), "/power/off", nil)
}

func (c *StripClient) SetRGB(r, g, b int) error {
	return c.SetColor(RGB{r, g, b})
}

func (c *StripClient) SetColor(rgb RGB) error {
	if err := validateRGB(rgb.R, rgb.G, rgb.B); err != nil {
		return err
	}
	return c.post(context.Background(), "/color", rgb)
}

func (c *StripClient) SetBrightness(percent int) error {
	if percent < 0 || percent > 100 {
		return fmt.Errorf("invalid strip brightness: %d", percent)
	}
	return c.post(context.Background(), "/brightness", map[string]int{"value": percent})
}

func (c *StripClient) SetSpeed(value int) error {
	if value < 0 || value > 255 {
		return fmt.Errorf("invalid strip speed: %d", value)
	}
	return c.post(context.Background(), "/speed", map[string]int{"value": value})
}

func (c *StripClient) SetMode(mode int) error {
	if mode < 0 || mode > 212 {
		return fmt.Errorf("invalid strip mode: %d", mode)
	}
	return c.post(context.Background(), "/mode", map[string]int{"mode": mode})
}

func (c *StripClient) SetTimer(d time.Duration) error {
	if d <= 0 {
		return fmt.Errorf("invalid strip timer duration: %s", d)
	}
	return c.post(context.Background(), "/timer", map[string]int{"seconds": int(d.Seconds())})
}

func (c *StripClient) CancelTimer() error {
	return c.do(context.Background(), http.MethodDelete, "/timer", nil, nil)
}

func (c *StripClient) Status() (StripStatus, error) {
	var status StripStatus
	if err := c.do(context.Background(), http.MethodGet, "/status", nil, &status); err != nil {
		return StripStatus{}, err
	}
	return status, nil
}

func (c *StripClient) post(ctx context.Context, path string, body any) error {
	return c.do(ctx, http.MethodPost, path, body, nil)
}

func (c *StripClient) do(ctx context.Context, method, path string, body any, out any) error {
	var payload io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal strip request: %w", err)
		}
		payload = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, payload)
	if err != nil {
		return fmt.Errorf("build strip request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("call strip service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("strip service %s %s: %s: %s", method, path, resp.Status, strings.TrimSpace(string(b)))
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode strip response: %w", err)
	}
	return nil
}

func parseRGB(text string) (int, int, int, error) {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "#") {
		hex := strings.TrimPrefix(text, "#")
		if len(hex) != 6 {
			return 0, 0, 0, fmt.Errorf("invalid hex rgb")
		}
		v, err := strconv.ParseUint(hex, 16, 32)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid hex rgb: %w", err)
		}
		return int((v >> 16) & 0xff), int((v >> 8) & 0xff), int(v & 0xff), nil
	}

	parts := strings.Fields(text)
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("rgb must contain three numbers")
	}
	r, err := parseRGBPart(parts[0])
	if err != nil {
		return 0, 0, 0, err
	}
	g, err := parseRGBPart(parts[1])
	if err != nil {
		return 0, 0, 0, err
	}
	b, err := parseRGBPart(parts[2])
	if err != nil {
		return 0, 0, 0, err
	}
	return r, g, b, nil
}

func parseRGBPart(text string) (int, error) {
	v, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("invalid rgb value %q: %w", text, err)
	}
	if v < 0 || v > 255 {
		return 0, fmt.Errorf("rgb value out of range: %d", v)
	}
	return v, nil
}

func validateRGB(r, g, b int) error {
	for _, v := range []int{r, g, b} {
		if v < 0 || v > 255 {
			return fmt.Errorf("rgb value out of range: %d", v)
		}
	}
	return nil
}

func stripTimerDuration(name string) (time.Duration, error) {
	durations := map[string]time.Duration{
		"Через 10 минут": 10 * time.Minute,
		"Через 30 минут": 30 * time.Minute,
		"Через 1 час":    time.Hour,
		"Через 2 часа":   2 * time.Hour,
	}
	duration, ok := durations[name]
	if !ok {
		return 0, fmt.Errorf("unknown strip timer: %s", name)
	}
	return duration, nil
}
