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
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// ErrDBVersionTableNotFound is returned when the database predates db_version.
var ErrDBVersionTableNotFound = errors.New("db_version table not found")

// DBVersion is the knowledge base's schema version, as reported on /status.
type DBVersion struct {
	PackageName   string `db:"package_name"`
	SchemaVersion string `db:"schema_version"`
	DBRelease     string `db:"db_release"`
	// CreatedAt is derived from DBRelease; see GetCurrentVersion.
	CreatedAt string `db:"-"`
}

// DBVersionModel reads the knowledge base's own version row.
//
// This replaces go-models' DBVersionModel, whose query selects a `created_at`
// column that the serving `db_version` does not have - the whole query fails on
// serving, so the API would silently stop reporting the version.
type DBVersionModel struct {
	db *sqlx.DB
}

func NewDBVersionModel(db *sqlx.DB) *DBVersionModel {
	return &DBVersionModel{db: db}
}

// GetCurrentVersion reads the serving db_version row.
//
// Serving records `db_release` rather than the mining `created_at`: a build
// stamp of the form YY.MM.DD.HH.MM (26.09.03.12.24 is 2026-09-03 12:24). The
// date is taken from it so /status keeps reporting when the knowledge base was
// built; a value in any other shape leaves CreatedAt empty rather than guessing.
func (m *DBVersionModel) GetCurrentVersion(ctx context.Context) (DBVersion, error) {
	if !m.tableExists(ctx) {
		return DBVersion{}, ErrDBVersionTableNotFound
	}
	var version DBVersion
	err := m.db.QueryRowxContext(ctx,
		"SELECT package_name, schema_version, db_release FROM db_version LIMIT 1").StructScan(&version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return DBVersion{}, nil
		}
		return DBVersion{}, fmt.Errorf("failed to query db_version table: %w", err)
	}
	version.CreatedAt = releaseDate(version.DBRelease)
	return version, nil
}

// releaseDate turns a YY.MM.DD.HH.MM build stamp into a calendar date, or
// returns empty when the value is not in that shape.
func releaseDate(release string) string {
	if release == "" {
		return ""
	}
	t, err := time.Parse("06.01.02.15.04", release)
	if err != nil {
		return ""
	}
	return t.Format(time.DateOnly)
}

// tableExists reports whether db_version is present, so a knowledge base that
// predates it degrades instead of erroring.
func (m *DBVersionModel) tableExists(ctx context.Context) bool {
	var exists int
	// information_schema is the portable spelling; sqlite, duckdb and postgres
	// all implement it, unlike sqlite_master.
	err := m.db.QueryRowxContext(ctx,
		m.db.Rebind("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?"),
		"db_version").Scan(&exists)
	return err == nil && exists > 0
}
