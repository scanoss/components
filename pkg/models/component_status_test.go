// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2025 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package models

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

func setupStatusModel(t *testing.T) (*sqlx.DB, *ComponentStatusModel) {
	t.Helper()
	if err := zlog.NewSugaredDevLogger(); err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	ctx := context.Background()
	db := sqliteSetup(t)
	if err := LoadTestSQLData(db, ctx, nil); err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	q := database.NewDBSelectContext(zlog.S, db, nil, false)
	return db, NewComponentStatusModel(ctx, zlog.S, q, db)
}

// firstVersionOf returns a version the fixture records for a component, so the
// version-level assertions do not hardcode one.
func firstVersionOf(t *testing.T, db *sqlx.DB, purlID string) string {
	t.Helper()
	var version string
	if err := db.Get(&version,
		"SELECT version FROM component_version WHERE purl_id = ? AND version <> '' LIMIT 1", purlID); err != nil {
		t.Fatalf("the fixture has no version for %s: %v", purlID, err)
	}
	return version
}

func TestGetComponentStatusByPurlID(t *testing.T) {
	db, model := setupStatusModel(t)
	defer CloseDB(db)
	defer zlog.SyncZap()

	t.Run("known component", func(t *testing.T) {
		status, err := model.GetComponentStatusByPurlID(tablestyleID)
		if err != nil {
			t.Fatalf("GetComponentStatusByPurlID error = %v", err)
		}
		if status.PurlName != tablestylePurl {
			t.Errorf("purl = %q, want %q", status.PurlName, tablestylePurl)
		}
		if status.Component == "" {
			t.Error("status came back without the component name")
		}
		if !status.Status.Valid || status.Status.String == "" {
			t.Error("status came back without a repository status")
		}
	})

	t.Run("unknown component", func(t *testing.T) {
		if _, err := model.GetComponentStatusByPurlID("00000000-0000-0000-0000-000000000000"); err == nil {
			t.Error("want an error for a component that is not there")
		}
	})

	t.Run("empty component is rejected", func(t *testing.T) {
		if _, err := model.GetComponentStatusByPurlID(""); err == nil {
			t.Error("want an error for an empty component")
		}
	})
}

func TestGetVersionStatusByPurlID(t *testing.T) {
	db, model := setupStatusModel(t)
	defer CloseDB(db)
	defer zlog.SyncZap()
	version := firstVersionOf(t, db, tablestyleID)

	t.Run("known release", func(t *testing.T) {
		status, err := model.GetVersionStatusByPurlID(tablestyleID, version)
		if err != nil {
			t.Fatalf("GetVersionStatusByPurlID error = %v", err)
		}
		if status.Version != version {
			t.Errorf("version = %q, want %q", status.Version, version)
		}
		if status.PurlName != tablestylePurl {
			t.Errorf("purl = %q, want %q", status.PurlName, tablestylePurl)
		}
	})

	t.Run("unknown release", func(t *testing.T) {
		if _, err := model.GetVersionStatusByPurlID(tablestyleID, "99.99.99"); err == nil {
			t.Error("want an error for a version that is not there")
		}
	})

	t.Run("empty component is rejected", func(t *testing.T) {
		if _, err := model.GetVersionStatusByPurlID("", version); err == nil {
			t.Error("want an error for an empty component")
		}
	})
}

// A missing table is a query failure, not an empty result.
func TestComponentStatusReportsQueryFailure(t *testing.T) {
	db, model := setupStatusModel(t)
	defer zlog.SyncZap()
	if _, err := db.Exec("DROP TABLE component_status;"); err != nil {
		t.Fatalf("failed to drop the table: %v", err)
	}
	if _, err := db.Exec("DROP TABLE component;"); err != nil {
		t.Fatalf("failed to drop the table: %v", err)
	}
	defer CloseDB(db)

	if _, err := model.GetComponentStatusByPurlID(tablestyleID); err == nil {
		t.Error("want an error once the tables are gone")
	}
}
