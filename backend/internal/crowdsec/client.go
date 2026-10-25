package crowdsec

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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