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

// The serving schema, used to build the database the tests run against.
//
// The fixtures under tests/ used to carry their own CREATE TABLE statements, each
// describing the table as that fixture's author imagined it. Keeping the schema
// here instead means every test is built from one description, and a query naming
// a column the real database does not have fails in CI rather than in production.
//
// This is a transcription of the serving schema, taken from the production DuckDB
// file with `SELECT sql FROM duckdb_tables()`. The types are the SQLite
// equivalents, since the tests run on SQLite: DuckDB UUID and DATE both become
// TEXT, which the queries already tolerate - they cast the UUIDs and the dates.
//
// Keep this in step with the real database. It is a transcription, so it can drift;
// when a query fails in production but passes here, suspect this file first.

package models

// testSchemaDDL is the schema every DB-backed test is built from.
const testSchemaDDL = `
CREATE TABLE db_version (package_name TEXT NOT NULL, schema_version TEXT NOT NULL, db_release TEXT NOT NULL);

-- Component identity. purl is the versionless purl string, and it carries the
-- purl type as its first segment: the mining "mines" table is gone.
CREATE TABLE component (purl TEXT, vendor TEXT, component TEXT, source_purl_id TEXT, purl_id TEXT);
CREATE INDEX idx_component_purl ON component (purl);
CREATE INDEX idx_component_purl_id ON component (purl_id);

-- Popularity, one optional row per component. These are the columns the search
-- ranking orders by; mining kept them on "projects".
CREATE TABLE component_metrics (
    purl_id TEXT, git_stars INTEGER, git_forks INTEGER,
    git_created_at TEXT, first_version_date TEXT, versions INTEGER
);
CREATE INDEX idx_component_metrics_purl_id ON component_metrics (purl_id);

CREATE TABLE component_status (
    purl_id TEXT, repository_status TEXT, first_indexed_date TEXT,
    last_indexed_date TEXT, status_change_date TEXT
);
CREATE INDEX idx_component_status_purl_id ON component_status (purl_id);

CREATE TABLE component_version (purl_id TEXT, version TEXT, version_name TEXT, version_semver TEXT);
CREATE INDEX idx_component_version_purl_id ON component_version (purl_id);

CREATE TABLE component_version_status (
    purl_id TEXT, version TEXT, repository_status TEXT,
    indexed_date TEXT, status_change_date TEXT
);
CREATE INDEX idx_component_version_status ON component_version_status (purl_id, version);

CREATE TABLE component_version_license (purl_id TEXT, version TEXT, license_id INTEGER, source INTEGER);
CREATE INDEX idx_component_version_license ON component_version_license (purl_id, version);

CREATE TABLE license (id INTEGER, long_name TEXT, SPDX TEXT, content TEXT);
CREATE INDEX idx_license_id ON license (id);

-- Whether a licence is SPDX-approved is no longer a stored flag: it is whether
-- the identifier appears here.
CREATE TABLE SPDX_license_data (
    SPDX_id TEXT, type TEXT, reference TEXT, is_deprecated BOOLEAN,
    details_url TEXT, reference_number INTEGER, name TEXT,
    is_osi_approved BOOLEAN, see_also TEXT
);
CREATE INDEX idx_spdx_license_data_id ON SPDX_license_data (SPDX_id);
`
