package crowdsec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	base string
	key  string
	hc   *http.Client
}

func New(base, key string) *Client {
	return &Client{
		base: stringsTrimSlash(base),
		key:  key,
		hc:   &http.Client{Timeout: 8 * time.Second},
	}
}

func stringsTrimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

type Decision struct {
	ID        int    `json:"id"`
	Origin    string `json:"origin"`
	Type      string `json:"type"`
	Scope     string `json:"scope"`
	Value     string `json:"value"`
	Duration  string `json:"duration"`
	Scenario  string `json:"scenario"`
	Simulated bool   `json:"simulated"`
}

func (c *Client) Decisions() ([]Decision, error) {
	if c.key == "" {
		return nil, fmt.Errorf("crowdsec key empty")
	}
	req, err := http.NewRequest(http.MethodGet, c.base+"/v1/decisions", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", c.key)
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("crowdsec %d: %s", resp.StatusCode, string(body))
	}
	if len(body) == 0 || string(body) == "null" {
		return []Decision{}, nil
	}
	var out []Decision
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) AddDecision(scope, value, typ, duration, reason string) error {
	if c.key == "" {
		return fmt.Errorf("crowdsec key empty")
	}
	if scope == "" {
		scope = "Ip"
	}
	if typ == "" {
		typ = "ban"
	}
	if duration == "" {
		duration = "87600h"
	}
	if reason == "" {
		reason = "the-gate"
	}
	body, _ := json.Marshal([]map[string]any{{
		"scope":    scope,
		"value":    value,
		"type":     typ,
		"duration": duration,
		"scenario": reason,
		"origin":   "the-gate",
	}})
	req, err := http.NewRequest(http.MethodPost, c.base+"/v1/decisions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", c.key)
	req.Header.Set("Content-Type", "application/json")
	return c.doNoBody(req)
}

func (c *Client) DeleteDecision(scope, value string) error {
	if c.key == "" {
		return fmt.Errorf("crowdsec key empty")
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("value required")
	}
	// Try common CrowdSec LAPI delete shapes (version differences).
	attempts := []url.Values{
		{"ip": {value}},
		{"value": {value}},
	}
	if scope != "" {
		attempts = append([]url.Values{{"scope": {scope}, "value": {value}}}, attempts...)
	}
	var last error
	for _, q := range attempts {
		req, err := http.NewRequest(http.MethodDelete, c.base+"/v1/decisions?"+q.Encode(), nil)
		if err != nil {
			last = err
			continue
		}
		req.Header.Set("X-Api-Key", c.key)
		if err := c.doNoBody(req); err != nil {
			last = err
			continue
		}
		return nil
	}
	if last == nil {
		last = fmt.Errorf("delete decision failed")
	}
	return last
}

func (c *Client) doNoBody(req *http.Request) error {
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("crowdsec %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
