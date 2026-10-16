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

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	"go.uber.org/zap"
)

type ComponentStatusModel struct {
	ctx context.Context
	s   *zap.SugaredLogger
	q   *database.DBQueryContext
	db  *sqlx.DB
}

// ComponentVersionStatus represents the status information for a specific version.
type ComponentVersionStatus struct {
	PurlName                string         `db:"purl_name"`
	Version                 string         `db:"version"`
	IndexedDate             sql.NullString `db:"indexed_date"`
	VersionStatus           sql.NullString `db:"version_status"`
	VersionStatusChangeDate sql.NullString `db:"version_status_change_date"`
}

// ComponentProjectStatus represents the status information for a component (ignoring version).
type ComponentProjectStatus struct {
	PurlName         string         `db:"purl_name"`
	Component        string         `db:"component"`
	FirstIndexedDate sql.NullString `db:"first_indexed_date"`
	LastIndexedDate  sql.NullString `db:"last_indexed_date"`
	Status           sql.NullString `db:"status"`
	StatusChangeDate sql.NullString `db:"status_change_date"`
}

// ComponentFullStatus combines version and project status information.
type ComponentFullStatus struct {
	ComponentVersionStatus
	ComponentProjectStatus
}

func NewComponentStatusModel(ctx context.Context, s *zap.SugaredLogger, q *database.DBQueryContext, db *sqlx.DB) *ComponentStatusModel {
	return &ComponentStatusModel{ctx: ctx, s: s, q: q, db: db}
}

// versionStatusQuery reads the status of one release.
//
// Mining kept this on `all_urls` (`version_status`, `indexed_date`,
// `version_status_change_date`) and needed a join to `mines` for the purl type.
// Serving has a dedicated `component_version_status` keyed by (purl_id, version),
// with the status column named `repository_status`.
const versionStatusQuery = `SELECT cvs.version AS version,
	CAST(cvs.indexed_date AS VARCHAR)        AS indexed_date,
	cvs.repository_status                    AS version_status,
	CAST(cvs.status_change_date AS VARCHAR)  AS version_status_change_date
FROM component_version_status cvs
WHERE cvs.purl_id = ? AND cvs.version = ?`

// componentStatusQuery reads the component-level status. Mining read it off
// `projects`; serving has `component_status`, one row per purl_id.
//
// No LIMIT, on purpose: with one, DuckDB answers this unindexed lookup with a
// streaming single-threaded scan, 1.4 s against 60 ms for the parallel scan it
// uses without. There is one row per purl_id; the caller takes the first.
const componentStatusQuery = `SELECT CAST(cs.first_indexed_date AS VARCHAR) AS first_indexed_date,
	CAST(cs.last_indexed_date AS VARCHAR)    AS last_indexed_date,
	cs.repository_status                     AS status,
	CAST(cs.status_change_date AS VARCHAR)   AS status_change_date
FROM component_status cs
WHERE cs.purl_id = ?`

// componentByPurlQuery reads the component row by its stored purl.
//
// The component row is read on its own, never joined, in every query of this
// package. `component(purl)` is the export's only index, and DuckDB does not use
// it once a join carries a purl_id filter across: it falls back to scanning all
// 32 million rows, about 2.5 s of CPU per call, which is what made a batch of
// status requests take seconds. Read alone by purl it takes a few milliseconds.
const componentByPurlQuery = `SELECT purl AS purl_name, component AS component,
	CAST(purl_id AS VARCHAR) AS purl_id
FROM component WHERE purl = ?`

// componentRow is the slice of a `component` row the models report.
type componentRow struct {
	PurlName  string         `db:"purl_name"`
	Component sql.NullString `db:"component"`
	PurlID    string         `db:"purl_id"`
}

// componentByPurl reads a component by its stored purl; nil when there is none.
func componentByPurl(ctx context.Context, q *database.DBQueryContext, db *sqlx.DB, purl string) (*componentRow, error) {
	var rows []componentRow
	if err := q.SelectContext(ctx, &rows, db.Rebind(componentByPurlQuery), purl); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// GetVersionStatusByPurlID gets the status of one release of a resolved component,
// given its purl_id and its purl as stored.
func (m *ComponentStatusModel) GetVersionStatusByPurlID(purlID, purl, version string) (*ComponentVersionStatus, error) {
	if len(purlID) == 0 || len(purl) == 0 {
		m.s.Error("Please specify a valid component to query")
		return nil, errors.New("please specify a valid component to query")
	}
	var results []ComponentVersionStatus
	err := m.q.SelectContext(m.ctx, &results, m.db.Rebind(versionStatusQuery), purlID, version)
	if err != nil {
		m.s.Errorf("Failed to query component status for %v version %v: %v", purlID, version, err)
		return nil, fmt.Errorf("failed to query component status: %w", err)
	}
	if len(results) == 0 {
		m.s.Warnf("No status found for %v version %v", purlID, version)
		return nil, errors.New("component version not found")
	}
	m.s.Debugf("Found status for %v version %v", purlID, version)
	results[0].PurlName = purl
	return &results[0], nil
}

// GetComponentStatusByPurl gets the component-level status of a resolved component,
// given its purl as stored.
func (m *ComponentStatusModel) GetComponentStatusByPurl(purl string) (*ComponentProjectStatus, error) {
	if len(purl) == 0 {
		m.s.Error("Please specify a valid component to query")
		return nil, errors.New("please specify a valid component to query")
	}
	comp, err := componentByPurl(m.ctx, m.q, m.db, purl)
	if err != nil {
		m.s.Errorf("Failed to query component %v: %v", purl, err)
		return nil, fmt.Errorf("failed to query component status: %w", err)
	}
	if comp == nil {
		m.s.Warnf("No status found for %v", purl)
		return nil, errors.New("component not found")
	}
	// A component with no status row still reports its name, as the LEFT JOIN this
	// replaces did.
	status := ComponentProjectStatus{PurlName: comp.PurlName, Component: comp.Component.String}
	var results []ComponentProjectStatus
	if err = m.q.SelectContext(m.ctx, &results, m.db.Rebind(componentStatusQuery), comp.PurlID); err != nil {
		m.s.Errorf("Failed to query component status for %v: %v", purl, err)
		return nil, fmt.Errorf("failed to query component status: %w", err)
	}
	if len(results) > 0 {
		status.FirstIndexedDate = results[0].FirstIndexedDate
		status.LastIndexedDate = results[0].LastIndexedDate
		status.Status = results[0].Status
		status.StatusChangeDate = results[0].StatusChangeDate
	}
	m.s.Debugf("Found status for %v", purl)
	return &status, nil
}
