package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ESP32Client struct {
	baseURL string
	client  *http.Client
}

func newESP32Client(baseURL string) *ESP32Client {
	return &ESP32Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *ESP32Client) Ping() error {
	return c.get(context.Background(), "/", nil)
}

func (c *ESP32Client) Red() error {
	return c.colorPath("/red", -1)
}

func (c *ESP32Client) Green() error {
	return c.colorPath("/green", -1)
}

func (c *ESP32Client) Blue() error {
	return c.colorPath("/blue", -1)
}

func (c *ESP32Client) Off() error {
	return c.get(context.Background(), "/off", nil)
}

func (c *ESP32Client) RedBrightness(value int) error {
	return c.colorPath("/red", value)
}

func (c *ESP32Client) SetRGB(r, g, b int) error {
	return c.SetRGBBrightness(r, g, b, -1)
}

func (c *ESP32Client) SetRGBBrightness(r, g, b, brightness int) error {
	if err := validateRGB(r, g, b); err != nil {
		return err
	}
	params := url.Values{
		"r": []string{strconv.Itoa(r)},
		"g": []string{strconv.Itoa(g)},
		"b": []string{strconv.Itoa(b)},
	}
	if brightness >= 0 {
		if brightness > 255 {
			return fmt.Errorf("invalid esp32 brightness: %d", brightness)
		}
		params.Set("brightness", strconv.Itoa(brightness))
	}
	return c.get(context.Background(), "/color", params)
}

func (c *ESP32Client) colorPath(path string, brightness int) error {
	params := url.Values{}
	if brightness >= 0 {
		if brightness > 255 {
			return fmt.Errorf("invalid esp32 brightness: %d", brightness)
		}
		params.Set("brightness", strconv.Itoa(brightness))
	}
	return c.get(context.Background(), path, params)
}

func (c *ESP32Client) get(ctx context.Context, path string, params url.Values) error {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("build esp32 request: %w", err)
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("build esp32 request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("call esp32: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("esp32 %s: %s: %s", path, resp.Status, strings.TrimSpace(string(b)))
	}
	return nil
}
