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

// The fixture's pkg:gem/tablestyle, a component with a healthy release list.
const (
	tablestylePurl = "pkg:gem/tablestyle"
	tablestyleID   = "bb7b1725-5229-54b2-949f-66143ae3e340"
)

func setupVersionsModel(t *testing.T) (*sqlx.DB, *ComponentVersionsModel) {
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
	return db, NewComponentVersionsModel(ctx, zlog.S, q, db)
}

func TestGetVersionsByPurlID(t *testing.T) {
	db, model := setupVersionsModel(t)
	defer CloseDB(db)
	defer zlog.SyncZap()

	t.Run("known component", func(t *testing.T) {
		versions, err := model.GetVersionsByPurlID(tablestyleID, 0)
		if err != nil {
			t.Fatalf("GetVersionsByPurlID error = %v", err)
		}
		if len(versions) == 0 {
			t.Fatal("no versions returned for a component the fixture carries")
		}
		for _, v := range versions {
			// The package-level aggregate row is not a release and must not appear.
			if v.Version == "" {
				t.Error("the empty aggregate version leaked into the release list")
			}
			if v.PurlName != tablestylePurl {
				t.Errorf("purl = %q, want %q", v.PurlName, tablestylePurl)
			}
			if v.Component == "" {
				t.Error("release came back without the component name")
			}
		}
	})

	t.Run("the limit bounds the rows", func(t *testing.T) {
		versions, err := model.GetVersionsByPurlID(tablestyleID, 3)
		if err != nil {
			t.Fatalf("GetVersionsByPurlID error = %v", err)
		}
		if len(versions) > 3 {
			t.Errorf("got %d rows, want at most 3", len(versions))
		}
	})

	t.Run("unknown component returns nothing", func(t *testing.T) {
		versions, err := model.GetVersionsByPurlID("00000000-0000-0000-0000-000000000000", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(versions) != 0 {
			t.Errorf("got %d rows, want none", len(versions))
		}
	})

	t.Run("an empty component is rejected", func(t *testing.T) {
		if _, err := model.GetVersionsByPurlID("", 0); err == nil {
			t.Error("want an error for an empty component")
		}
	})
}

// A missing table is a query failure, not an empty result.
func TestGetVersionsByPurlIDReportsQueryFailure(t *testing.T) {
	db, model := setupVersionsModel(t)
	defer zlog.SyncZap()
	if _, err := db.Exec("DROP TABLE component_version;"); err != nil {
		t.Fatalf("failed to drop the table: %v", err)
	}
	defer CloseDB(db)

	if _, err := model.GetVersionsByPurlID(tablestyleID, 0); err == nil {
		t.Error("want an error once the table is gone")
	}
}
