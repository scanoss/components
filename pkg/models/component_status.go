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
const versionStatusQuery = `SELECT c.purl        AS purl_name,
	cvs.version                              AS version,
	CAST(cvs.indexed_date AS VARCHAR)        AS indexed_date,
	cvs.repository_status                    AS version_status,
	CAST(cvs.status_change_date AS VARCHAR)  AS version_status_change_date
FROM component_version_status cvs
INNER JOIN component c ON c.purl_id = cvs.purl_id
WHERE cvs.purl_id = ? AND cvs.version = ?`

// componentStatusQuery reads the component-level status. Mining read it off
// `projects`; serving has `component_status`, one row per purl_id.
const componentStatusQuery = `SELECT c.purl      AS purl_name,
	c.component                              AS component,
	CAST(cs.first_indexed_date AS VARCHAR)   AS first_indexed_date,
	CAST(cs.last_indexed_date AS VARCHAR)    AS last_indexed_date,
	cs.repository_status                     AS status,
	CAST(cs.status_change_date AS VARCHAR)   AS status_change_date
FROM component c
LEFT JOIN component_status cs ON cs.purl_id = c.purl_id
WHERE c.purl_id = ?
LIMIT 1`

// GetVersionStatusByPurlID gets the status of one release of a resolved component.
func (m *ComponentStatusModel) GetVersionStatusByPurlID(purlID, version string) (*ComponentVersionStatus, error) {
	if len(purlID) == 0 {
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
	return &results[0], nil
}

// GetComponentStatusByPurlID gets the component-level status of a resolved component.
func (m *ComponentStatusModel) GetComponentStatusByPurlID(purlID string) (*ComponentProjectStatus, error) {
	if len(purlID) == 0 {
		m.s.Error("Please specify a valid component to query")
		return nil, errors.New("please specify a valid component to query")
	}
	var results []ComponentProjectStatus
	err := m.q.SelectContext(m.ctx, &results, m.db.Rebind(componentStatusQuery), purlID)
	if err != nil {
		m.s.Errorf("Failed to query component status for %v: %v", purlID, err)
		return nil, fmt.Errorf("failed to query component status: %w", err)
	}
	if len(results) == 0 {
		m.s.Warnf("No status found for %v", purlID)
		return nil, errors.New("component not found")
	}
	m.s.Debugf("Found status for %v", purlID)
	return &results[0], nil
}
