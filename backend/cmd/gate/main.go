// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package main

import (
	"log"
	"os"

	"github.com/overthecloud/the-gate/internal/api"
	"github.com/overthecloud/the-gate/internal/config"
	"github.com/overthecloud/the-gate/internal/store"
	"github.com/overthecloud/the-gate/internal/traefiksync"
)

func main() {
	cfg := config.FromEnv()
	db, err := store.Open(cfg.DBURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureAdmin(db, cfg.AdminUser, cfg.AdminPassword); err != nil {
		log.Fatalf("admin: %v", err)
	}

	syncer := traefiksync.New(cfg.TraefikDynamicDir, cfg.BouncerKey, cfg.AuthentikForwardURL)
	if err := syncer.WriteBaseMiddlewares(); err != nil {
		log.Fatalf("middlewares: %v", err)
	}
	apps, err := store.ListApplications(db)
	if err != nil {
		log.Fatalf("list apps: %v", err)
	}
	if err := syncer.SyncApplications(apps); err != nil {
		log.Fatalf("sync apps: %v", err)
	}

	r := api.NewRouter(cfg, db, syncer)
	log.Printf("The Gate API listening on %s (auth_mode=%s oidc=%v)", cfg.Listen, cfg.AuthMode, cfg.OIDCEnabled())
	if err := r.Run(cfg.Listen); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}