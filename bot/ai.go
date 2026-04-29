package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type aiResult struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
	Text  string  `json:"text"`
}

func postJSON(url string, payload any, out any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ai status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if out == nil || len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, out)
}

func aiAdd(aiURL string, n Note) error {
	return postJSON(aiURL+"/add", map[string]string{"id": n.ID, "text": n.Text}, nil)
}

func aiDelete(aiURL, id string) error {
	return postJSON(aiURL+"/delete", map[string]string{"id": id}, nil)
}

func aiSearch(aiURL, query string) ([]aiResult, error) {
	var res []aiResult
	if err := postJSON(aiURL+"/search", map[string]string{"query": query}, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func aiHealth(aiURL string) error {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(aiURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ai health status: %d", resp.StatusCode)
	}
	return nil
}

func waitForAI(aiURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := aiHealth(aiURL); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(2 * time.Second)
	}
	if lastErr == nil {
		lastErr = errors.New("ai health check timeout")
	}
	return fmt.Errorf("ai service unavailable after %s: %w", timeout.String(), lastErr)
}

func reindexToAI(store *NotesStore, aiURL string) error {
	for _, n := range store.All() {
		if err := aiAdd(aiURL, n); err != nil {
			return err
		}
	}
	return nil
}
