// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2026 SCANOSS.COM
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
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	scanossdb "scanoss.com/components/pkg/db"
)

// duckdbSetup loads the test fixtures into an in-memory DuckDB with the serving
// DATE columns restored. The SQLite fixtures keep dates as text, which hides how
// the DuckDB driver hands a DATE over: as a time.Time, that sql.NullString
// renders as "2013-08-26T00:00:00Z".
func duckdbSetup(t *testing.T) *sqlx.DB {
	t.Helper()
	if err := zlog.NewSugaredDevLogger(); err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	db, err := sqlx.Connect(scanossdb.DriverDuckDB, "")
	if err != nil {
		t.Fatalf("failed to open an in-memory DuckDB: %v", err)
	}
	db.SetMaxOpenConns(1)
	if err = LoadTestSchema(db, nil, nil); err != nil {
		t.Fatalf("failed to load the test schema: %v", err)
	}
	// The indexes block an ALTER of the column type, so the two status tables are
	// recreated with the serving types instead.
	for _, stmt := range []string{
		"DROP INDEX idx_component_status_purl_id",
		"DROP TABLE component_status",
		`CREATE TABLE component_status (purl_id TEXT, repository_status TEXT,
			first_indexed_date DATE, last_indexed_date DATE, status_change_date DATE)`,
		"DROP INDEX idx_component_version_status",
		"DROP TABLE component_version_status",
		`CREATE TABLE component_version_status (purl_id TEXT, version TEXT,
			repository_status TEXT, indexed_date DATE, status_change_date DATE)`,
	} {
		if _, err = db.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	if err = LoadTestSQLData(db, nil, nil); err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	return db
}

// The dates keep the format the mining-backed service returned: release dates in
// the version list and every status date bare, but a release's indexed_date as an
// RFC 3339 timestamp.
func TestServingDateFormatsOnDuckDB(t *testing.T) {
	db := duckdbSetup(t)
	defer CloseDB(db)
	defer zlog.SyncZap()
	ctx := context.Background()
	q := database.NewDBSelectContext(zlog.S, db, nil, false)

	versions, err := NewComponentVersionsModel(ctx, zlog.S, q, db).GetVersionsByPurlID(tablestyleID, tablestylePurl, 0)
	if err != nil {
		t.Fatalf("GetVersionsByPurlID error = %v", err)
	}
	if len(versions) == 0 {
		t.Fatal("no versions returned for a component the fixture carries")
	}
	for _, v := range versions {
		if _, err := time.Parse(time.DateOnly, v.Date.String); v.Date.Valid && err != nil {
			t.Errorf("version %s date = %q, want a bare date (YYYY-MM-DD)", v.Version, v.Date.String)
		}
	}

	status, err := NewComponentStatusModel(ctx, zlog.S, q, db).GetVersionStatusByPurlID(tablestyleID, tablestylePurl, "0.0.10")
	if err != nil {
		t.Fatalf("GetVersionStatusByPurlID error = %v", err)
	}
	if got := status.IndexedDate.String; got != "2013-08-26T00:00:00Z" {
		t.Errorf("indexed_date = %q, want 2013-08-26T00:00:00Z", got)
	}
	if got := status.VersionStatusChangeDate.String; got != "2013-08-26" {
		t.Errorf("status_change_date = %q, want 2013-08-26", got)
	}
}
