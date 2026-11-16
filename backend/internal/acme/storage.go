// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package acme

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var mu sync.Mutex

type CertInfo struct {
	Resolver   string    `json:"resolver"`
	Main       string    `json:"main"`
	SANs       []string  `json:"sans"`
	Domains    []string  `json:"domains"`
	Issuer     string    `json:"issuer"`
	NotBefore  time.Time `json:"not_before"`
	NotAfter   time.Time `json:"not_after"`
	DaysLeft   int       `json:"days_left"`
	AppID      string    `json:"app_id,omitempty"`
	AppName    string    `json:"app_name,omitempty"`
	AppHost    string    `json:"app_host,omitempty"`
}

type domainBlock struct {
	Main string   `json:"main"`
	SANs []string `json:"sans"`
}

type certEntry struct {
	Domain      domainBlock `json:"domain"`
	Certificate string      `json:"certificate"`
	Key         string      `json:"key"`
	Store       string      `json:"Store"`
}

type resolverBlock struct {
	Account      json.RawMessage `json:"Account"`
	Certificates []certEntry     `json:"Certificates"`
}

func List(path string) ([]CertInfo, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []CertInfo{}, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return []CertInfo{}, nil
	}
	var root map[string]resolverBlock
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, err
	}
	out := make([]CertInfo, 0)
	for resolver, block := range root {
		for _, c := range block.Certificates {
			info, err := parseCert(resolver, c)
			if err != nil {
				continue
			}
			out = append(out, info)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Main < out[j].Main
	})
	return out, nil
}

func parseCert(resolver string, c certEntry) (CertInfo, error) {
	der, err := decodeCert(c.Certificate)
	if err != nil {
		return CertInfo{}, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return CertInfo{}, err
	}
	main := strings.ToLower(strings.TrimSpace(c.Domain.Main))
	if main == "" && len(cert.DNSNames) > 0 {
		main = strings.ToLower(cert.DNSNames[0])
	}
	sans := make([]string, 0, len(c.Domain.SANs))
	for _, s := range c.Domain.SANs {
		s = strings.ToLower(strings.TrimSpace(s))
		if s != "" && s != main {
			sans = append(sans, s)
		}
	}
	if len(sans) == 0 {
		for _, d := range cert.DNSNames {
			d = strings.ToLower(d)
			if d != "" && d != main {
				sans = append(sans, d)
			}
		}
	}
	domains := append([]string{}, main)
	domains = append(domains, sans...)
	days := int(time.Until(cert.NotAfter).Hours() / 24)
	return CertInfo{
		Resolver:  resolver,
		Main:      main,
		SANs:      sans,
		Domains:   domains,
		Issuer:    cert.Issuer.CommonName,
		NotBefore: cert.NotBefore.UTC(),
		NotAfter:  cert.NotAfter.UTC(),
		DaysLeft:  days,
	}, nil
}

func decodeCert(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("empty certificate")
	}
	if strings.Contains(raw, "BEGIN CERTIFICATE") {
		block, _ := pem.Decode([]byte(raw))
		if block == nil {
			return nil, errors.New("invalid pem")
		}
		return block.Bytes, nil
	}
	b, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	// Traefik may store PEM as base64
	if pemBlock, _ := pem.Decode(b); pemBlock != nil {
		return pemBlock.Bytes, nil
	}
	return b, nil
}

// ForceRenew removes matching certificate entries so Traefik re-issues via ACME.
func ForceRenew(path, domain string) (removed int, err error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return 0, errors.New("domain required")
	}
	mu.Lock()
	defer mu.Unlock()

	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return 0, errors.New("acme storage empty")
	}
	var root map[string]resolverBlock
	if err := json.Unmarshal(raw, &root); err != nil {
		return 0, err
	}
	changed := false
	for name, block := range root {
		filtered := make([]certEntry, 0, len(block.Certificates))
		for _, c := range block.Certificates {
			if certMatches(c, domain) {
				removed++
				changed = true
				continue
			}
			filtered = append(filtered, c)
		}
		block.Certificates = filtered
		root[name] = block
	}
	if !changed {
		return 0, errors.New("certificate not found for domain")
	}
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return removed, err
	}
	return removed, os.WriteFile(path, out, 0o600)
}

func certMatches(c certEntry, domain string) bool {
	if strings.EqualFold(c.Domain.Main, domain) {
		return true
	}
	for _, s := range c.Domain.SANs {
		if strings.EqualFold(s, domain) {
			return true
		}
	}
	return false
}
