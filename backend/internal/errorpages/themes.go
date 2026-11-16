// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package errorpages

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Theme struct {
	ID          string `json:"id"`
	NameEN      string `json:"name_en"`
	NameIT      string `json:"name_it"`
	Ink         string `json:"ink"`
	Muted       string `json:"muted"`
	Brand       string `json:"brand"`
	Background  string `json:"background"`
	AccentGlow  string `json:"accent_glow"` // rgba for radial gradient
	Preview     string `json:"preview"`     // brand hex for UI swatch
}

var Themes = []Theme{
	{ID: "forest", NameEN: "Forest (The Gate)", NameIT: "Foresta (The Gate)", Ink: "#1a1f1c", Muted: "#5c675f", Brand: "#0f3d2e", Background: "#f3efe6", AccentGlow: "rgba(31,111,84,.12)", Preview: "#0f3d2e"},
	{ID: "ocean", NameEN: "Ocean Blue", NameIT: "Blu oceano", Ink: "#0f172a", Muted: "#64748b", Brand: "#0369a1", Background: "#f0f9ff", AccentGlow: "rgba(3,105,161,.14)", Preview: "#0369a1"},
	{ID: "slate", NameEN: "Slate Neutral", NameIT: "Ardesia neutra", Ink: "#111827", Muted: "#6b7280", Brand: "#374151", Background: "#f9fafb", AccentGlow: "rgba(55,65,81,.12)", Preview: "#374151"},
	{ID: "teal", NameEN: "Safe Teal", NameIT: "Teal sicurezza", Ink: "#134e4a", Muted: "#5f8a86", Brand: "#0d9488", Background: "#f0fdfa", AccentGlow: "rgba(13,148,136,.14)", Preview: "#0d9488"},
	{ID: "coral", NameEN: "Alert Coral", NameIT: "Corallo allerta", Ink: "#1c1917", Muted: "#78716c", Brand: "#c2410c", Background: "#fff7ed", AccentGlow: "rgba(194,65,12,.14)", Preview: "#c2410c"},
	{ID: "violet", NameEN: "Soft Violet", NameIT: "Viola soft", Ink: "#1e1b4b", Muted: "#6b7280", Brand: "#6d28d9", Background: "#f5f3ff", AccentGlow: "rgba(109,40,217,.12)", Preview: "#6d28d9"},
	{ID: "midnight", NameEN: "Midnight Dark", NameIT: "Mezzanotte scuro", Ink: "#e5e7eb", Muted: "#9ca3af", Brand: "#38bdf8", Background: "#0b1220", AccentGlow: "rgba(56,189,248,.16)", Preview: "#0b1220"},
	{ID: "sand", NameEN: "Warm Sand", NameIT: "Sabbia calda", Ink: "#292524", Muted: "#78716c", Brand: "#a16207", Background: "#faf6f1", AccentGlow: "rgba(161,98,7,.12)", Preview: "#a16207"},
}

func ThemeByID(id string) (Theme, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, t := range Themes {
		if t.ID == id {
			return t, true
		}
	}
	return Theme{}, false
}

var (
	reRoot = regexp.MustCompile(`(?is):root\s*\{[^}]*\}`)
	reBodyBg = regexp.MustCompile(`(?is)background:\s*radial-gradient\([^;]+\),\s*var\(--bg\)`)
)

// ApplyTheme rewrites CSS variables in error-page HTML to the chosen palette.
func ApplyTheme(html string, theme Theme) string {
	root := fmt.Sprintf(
		`:root { color-scheme: %s; --ink:%s; --muted:%s; --brand:%s; --bg:%s; }`,
		schemeFor(theme),
		theme.Ink,
		theme.Muted,
		theme.Brand,
		theme.Background,
	)
	if reRoot.MatchString(html) {
		html = reRoot.ReplaceAllString(html, root)
	} else if i := strings.Index(strings.ToLower(html), "<style>"); i >= 0 {
		j := i + len("<style>")
		html = html[:j] + "\n    " + root + "\n" + html[j:]
	}

	grad := fmt.Sprintf(
		`background: radial-gradient(900px 500px at 10%% -10%%, %s, transparent 55%%), var(--bg)`,
		theme.AccentGlow,
	)
	if reBodyBg.MatchString(html) {
		html = reBodyBg.ReplaceAllString(html, grad)
	}
	return html
}

func schemeFor(t Theme) string {
	if t.ID == "midnight" {
		return "dark"
	}
	return "light"
}

// ApplyThemeAll loads every template, applies theme, saves with tip.
func ApplyThemeAll(dir, themeID, tip string) error {
	theme, ok := ThemeByID(themeID)
	if !ok {
		return fmt.Errorf("unknown theme: %s", themeID)
	}
	entries, err := listCodes(dir)
	if err != nil {
		return err
	}
	for _, code := range entries {
		html, err := LoadTemplate(dir, code)
		if err != nil {
			continue
		}
		html = ApplyTheme(html, theme)
		if err := SaveTemplate(dir, code, html, tip); err != nil {
			return err
		}
	}
	return nil
}

func listCodes(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	seen := map[string]bool{}
	var codes []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
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
		codes = append(codes, code)
	}
	return codes, nil
}
