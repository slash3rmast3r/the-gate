// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package traffic

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/overthecloud/the-gate/internal/applogs"
	"github.com/overthecloud/the-gate/internal/geoip"
	"github.com/overthecloud/the-gate/internal/models"
)

type AppTraffic struct {
	ApplicationID string `json:"application_id"`
	Name          string `json:"name"`
	Host          string `json:"host"`
	Requests      int    `json:"requests"`
	Errors        int    `json:"errors"`
	Blocked       int    `json:"blocked"`
	UniqueClients int    `json:"unique_clients"`
	TopCountry    string `json:"top_country"`
}

type ClientTraffic struct {
	IP       string `json:"ip"`
	Country  string `json:"country"`
	Requests int    `json:"requests"`
	Errors   int    `json:"errors"`
	Blocked  int    `json:"blocked"`
}

type Summary struct {
	WindowMinutes int             `json:"window_minutes"`
	Requests      int             `json:"requests"`
	Errors        int             `json:"errors"`
	Blocked       int             `json:"blocked"`
	Applications  []AppTraffic    `json:"applications"`
	Clients       []ClientTraffic `json:"clients"`
	Countries     map[string]int  `json:"countries"`
}

type Resolver interface {
	Country(ip string) string
}

func Analyze(apps []models.Application, entries []applogs.Entry, r Resolver, minutes int) Summary {
	if minutes < 1 {
		minutes = 60
	}
	if r == nil {
		r = geoip.New("")
	}
	cutoff := time.Now().UTC().Add(-time.Duration(minutes) * time.Minute)
	appByHost := map[string]models.Application{}
	for _, a := range apps {
		appByHost[strings.ToLower(a.Host)] = a
	}

	out := Summary{
		WindowMinutes: minutes,
		Countries:     map[string]int{},
		Applications:  []AppTraffic{},
		Clients:       []ClientTraffic{},
	}
	appStats := map[string]*AppTraffic{}
	appCountries := map[string]map[string]int{}
	clients := map[string]*ClientTraffic{}
	seenByApp := map[string]map[string]bool{}

	for _, e := range entries {
		if !withinWindow(e.Time, cutoff) {
			continue
		}
		host := strings.ToLower(stripPort(e.Host))
		a, ok := appByHost[host]
		if !ok {
			continue
		}
		key := a.ID.String()
		if appStats[key] == nil {
			appStats[key] = &AppTraffic{ApplicationID: key, Name: a.Name, Host: a.Host}
			appCountries[key] = map[string]int{}
			seenByApp[key] = map[string]bool{}
		}
		ip := stripPort(e.Client)
		country := r.Country(ip)
		blocked := e.Status == 401 || e.Status == 403 || e.Status == 429
		isErr := e.Status >= 500

		out.Requests++
		appStats[key].Requests++
		if isErr {
			out.Errors++
			appStats[key].Errors++
		}
		if blocked {
			out.Blocked++
			appStats[key].Blocked++
		}
		if ip != "" {
			seenByApp[key][ip] = true
			if clients[ip] == nil {
				clients[ip] = &ClientTraffic{IP: ip, Country: country}
			}
			clients[ip].Requests++
			if isErr {
				clients[ip].Errors++
			}
			if blocked {
				clients[ip].Blocked++
			}
		}
		appCountries[key][country]++
		out.Countries[country]++
	}

	for key, stat := range appStats {
		stat.UniqueClients = len(seenByApp[key])
		stat.TopCountry = topCountry(appCountries[key])
		out.Applications = append(out.Applications, *stat)
	}
	for _, c := range clients {
		out.Clients = append(out.Clients, *c)
	}
	sort.Slice(out.Applications, func(i, j int) bool { return out.Applications[i].Requests > out.Applications[j].Requests })
	sort.Slice(out.Clients, func(i, j int) bool { return out.Clients[i].Requests > out.Clients[j].Requests })
	if len(out.Clients) > 50 {
		out.Clients = out.Clients[:50]
	}
	return out
}

func FormatMinutes(minutes int) string {
	if minutes < 60 {
		return strconv.Itoa(minutes) + "m"
	}
	if minutes%1440 == 0 {
		return strconv.Itoa(minutes/1440) + "d"
	}
	if minutes%60 == 0 {
		return strconv.Itoa(minutes/60) + "h"
	}
	return strconv.Itoa(minutes) + "m"
}

func withinWindow(raw string, cutoff time.Time) bool {
	if raw == "" {
		return true
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05Z", "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, raw); err == nil {
			return !ts.Before(cutoff)
		}
	}
	return true
}

func stripPort(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.LastIndex(s, ":"); i > 0 && strings.Count(s, ":") == 1 {
		return s[:i]
	}
	return s
}

func topCountry(m map[string]int) string {
	top := "Unknown"
	count := -1
	for c, n := range m {
		if n > count {
			top = c
			count = n
		}
	}
	return top
}
