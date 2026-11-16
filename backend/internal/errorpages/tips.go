// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package errorpages

import (
	"os"
	"path/filepath"
	"strings"
)

const TipPlaceholder = "{{TIP_MESSAGE}}"

func EnsureTipPlaceholder(html string) string {
	if strings.Contains(html, TipPlaceholder) {
		return html
	}
	inject := "\n    <p class=\"tip\">" + TipPlaceholder + "</p>\n    "
	lower := strings.ToLower(html)
	if i := strings.LastIndex(lower, "</main>"); i >= 0 {
		return html[:i] + inject + html[i:]
	}
	if i := strings.LastIndex(lower, "</body>"); i >= 0 {
		return html[:i] + inject + html[i:]
	}
	return html + "\n<p class=\"tip\">" + TipPlaceholder + "</p>\n"
}

func ApplyTip(html, tip string) string {
	tip = strings.TrimSpace(tip)
	if tip == "" {
		tip = "Access is now checked for threats."
	}
	tip = strings.ReplaceAll(tip, "&", "&amp;")
	tip = strings.ReplaceAll(tip, "<", "&lt;")
	tip = strings.ReplaceAll(tip, ">", "&gt;")
	tip = strings.ReplaceAll(tip, "\"", "&quot;")
	return strings.ReplaceAll(html, TipPlaceholder, tip)
}

// SaveTemplate stores the editable template (with placeholder) and renders the served HTML.
func SaveTemplate(dir, code, html, tip string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	html = EnsureTipPlaceholder(html)
	tpl := filepath.Join(dir, code+".html.tpl")
	main := filepath.Join(dir, code+".html")
	if err := os.WriteFile(tpl, []byte(html), 0o644); err != nil {
		return err
	}
	return os.WriteFile(main, []byte(ApplyTip(html, tip)), 0o644)
}

// LoadTemplate prefers .tpl (source with placeholder), falls back to .html.
func LoadTemplate(dir, code string) (string, error) {
	tpl := filepath.Join(dir, code+".html.tpl")
	if b, err := os.ReadFile(tpl); err == nil {
		return string(b), nil
	}
	b, err := os.ReadFile(filepath.Join(dir, code+".html"))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// RenderAll re-applies tip message to all error page templates.
func RenderAll(dir, tip string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	seen := map[string]bool{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			continue
		}
		code := ""
		switch {
		case strings.HasSuffix(name, ".html.tpl"):
			code = strings.TrimSuffix(name, ".html.tpl")
		case strings.HasSuffix(name, ".html"):
			code = strings.TrimSuffix(name, ".html")
		default:
			continue
		}
		if seen[code] {
			continue
		}
		seen[code] = true
		html, err := LoadTemplate(dir, code)
		if err != nil {
			continue
		}
		_ = SaveTemplate(dir, code, html, tip)
	}
	return nil
}
