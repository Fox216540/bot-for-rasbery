package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type OpenMeteoClient struct {
	baseURL string
	lat     float64
	lon     float64
	client  *http.Client
}

func newOpenMeteoClient(baseURL string, lat, lon float64) *OpenMeteoClient {
	return &OpenMeteoClient{
		baseURL: baseURL,
		lat:     lat,
		lon:     lon,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *OpenMeteoClient) SunriseSunset(ctx context.Context) (time.Time, time.Time, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("build meteo request: %w", err)
	}
	q := u.Query()
	q.Set("latitude", strconv.FormatFloat(c.lat, 'f', -1, 64))
	q.Set("longitude", strconv.FormatFloat(c.lon, 'f', -1, 64))
	q.Set("daily", "sunrise,sunset")
	q.Set("timezone", "auto")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("build meteo request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("call open-meteo: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return time.Time{}, time.Time{}, fmt.Errorf("open-meteo status: %s", resp.Status)
	}

	var payload struct {
		Timezone         string `json:"timezone"`
		UTCOffsetSeconds int    `json:"utc_offset_seconds"`
		Daily            struct {
			Sunrise []string `json:"sunrise"`
			Sunset  []string `json:"sunset"`
		} `json:"daily"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("decode open-meteo response: %w", err)
	}
	if len(payload.Daily.Sunrise) == 0 || len(payload.Daily.Sunset) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("open-meteo response missing sunrise or sunset")
	}
	loc := time.FixedZone(payload.Timezone, payload.UTCOffsetSeconds)
	sunrise, err := parseMeteoTime(payload.Daily.Sunrise[0], loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse sunrise: %w", err)
	}
	sunset, err := parseMeteoTime(payload.Daily.Sunset[0], loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse sunset: %w", err)
	}
	return sunrise, sunset, nil
}

func parseMeteoTime(value string, loc *time.Location) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	return time.ParseInLocation("2006-01-02T15:04", value, loc)
}

type StatusLightService struct {
	esp32           *ESP32Client
	meteo           *OpenMeteoClient
	dayBrightness   int
	nightBrightness int
	now             func() time.Time
}

func newStatusLightService(esp32 *ESP32Client, meteo *OpenMeteoClient, dayBrightness, nightBrightness int) *StatusLightService {
	return &StatusLightService{
		esp32:           esp32,
		meteo:           meteo,
		dayBrightness:   dayBrightness,
		nightBrightness: nightBrightness,
		now:             time.Now,
	}
}

func (s *StatusLightService) Set(ctx context.Context, rgb RGB) (int, error) {
	brightness, err := s.Brightness(ctx)
	if err != nil {
		return 0, err
	}
	if err := s.esp32.SetRGBBrightness(rgb.R, rgb.G, rgb.B, brightness); err != nil {
		return 0, err
	}
	return brightness, nil
}

func (s *StatusLightService) Brightness(ctx context.Context) (int, error) {
	sunrise, sunset, err := s.meteo.SunriseSunset(ctx)
	if err != nil {
		return 0, err
	}
	now := s.now()
	if now.Before(sunrise) || !now.Before(sunset) {
		return s.nightBrightness, nil
	}
	return s.dayBrightness, nil
}
