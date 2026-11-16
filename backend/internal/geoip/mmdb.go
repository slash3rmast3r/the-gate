// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package geoip

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

type mmdbCountry struct {
	db *geoip2.Reader
}

func openMMDB(path string) (countryDB, error) {
	db, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &mmdbCountry{db: db}, nil
}

func (m *mmdbCountry) Country(ip net.IP) (string, error) {
	rec, err := m.db.Country(ip)
	if err != nil {
		return "", err
	}
	if rec.Country.IsoCode != "" {
		return rec.Country.IsoCode, nil
	}
	if rec.RegisteredCountry.IsoCode != "" {
		return rec.RegisteredCountry.IsoCode, nil
	}
	if rec.Continent.Code != "" {
		return rec.Continent.Code, nil
	}
	return "", nil
}

func (m *mmdbCountry) Close() error {
	return m.db.Close()
}
