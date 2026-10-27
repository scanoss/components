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

const defaultMaxVersionLimit = 50

// ComponentVersionsModel lists a component's releases. It was AllURLsModel, over
// the mining `all_urls` table; serving keeps the releases in `component_version`
// and their licences in `component_version_license`, both keyed by purl_id.
type ComponentVersionsModel struct {
	ctx context.Context
	s   *zap.SugaredLogger
	q   *database.DBQueryContext
	db  *sqlx.DB
}

// AllURL is one (version, licence) row. The name is kept because the response
// DTOs are built from it; there are no URLs involved any more.
type AllURL struct {
	Version   string         `db:"version"`
	Component string         `db:"-"`
	License   string         `db:"license"`
	LicenseID string         `db:"license_id"`
	IsSpdx    bool           `db:"is_spdx"`
	PurlName  string         `db:"-"`
	Date      sql.NullString `db:"date"`
	URL       string         `db:"-"`
}

func NewComponentVersionsModel(ctx context.Context, s *zap.SugaredLogger, q *database.DBQueryContext, db *sqlx.DB) *ComponentVersionsModel {
	return &ComponentVersionsModel{ctx: ctx, s: s, q: q, db: db}
}

// versionsByPurlIDQuery lists a component's releases newest-first, with one row
// per (version, licence).
//
// The mining query read `all_urls` joined to `mines` and `licenses`, ordered by
// the row's own date. Serving splits that: the release list is
// `component_version`, the date comes from `component_version_status.indexed_date`,
// and the licence from `component_version_license` through `license`. Whether a
// licence is SPDX-approved is no longer a stored flag - it is whether the SPDX
// identifier appears in `SPDX_license_data`.
//
// The empty version is excluded: it is the package-level aggregate row, not a
// release.
const versionsByPurlIDQuery = `SELECT cv.version                AS version,
	cvs.indexed_date                         AS date,
	COALESCE(l.long_name, '')                AS license,
	COALESCE(l."SPDX", '')                   AS license_id,
	CASE WHEN sld."SPDX_id" IS NOT NULL THEN 1 ELSE 0 END AS is_spdx
FROM component_version cv
LEFT JOIN component_version_status cvs ON cvs.purl_id = cv.purl_id AND cvs.version = cv.version
LEFT JOIN component_version_license cvl ON cvl.purl_id = cv.purl_id AND cvl.version = cv.version
LEFT JOIN license l ON l.id = cvl.license_id
LEFT JOIN "SPDX_license_data" sld ON sld."SPDX_id" = l."SPDX"
WHERE cv.purl_id = ? AND cv.version <> ''
ORDER BY date DESC NULLS LAST
LIMIT ?`

// GetVersionsByPurlID lists the releases of an already-resolved component, given
// its purl_id and its purl as stored. The component's name is read separately, by
// purl (see componentByPurlQuery).
func (m *ComponentVersionsModel) GetVersionsByPurlID(purlID, purl string, limit int) ([]AllURL, error) {
	if len(purlID) == 0 || len(purl) == 0 {
		m.s.Error("Please specify a valid component to query")
		return nil, errors.New("please specify a valid component to query")
	}
	if limit <= 0 {
		limit = defaultMaxVersionLimit
	}
	comp, err := componentByPurl(m.ctx, m.q, m.db, purl)
	if errors.Is(err, errNoComponent) {
		return []AllURL{}, nil
	}
	if err != nil {
		m.s.Errorf("Failed to query component %v: %v", purl, err)
		return nil, fmt.Errorf("failed to query the component versions: %w", err)
	}
	var versions []AllURL
	err = m.q.SelectContext(m.ctx, &versions, m.db.Rebind(versionsByPurlIDQuery), purlID, limit)
	if err != nil {
		m.s.Errorf("Failed to query component versions for %v: %v", purlID, err)
		return nil, fmt.Errorf("failed to query the component versions: %w", err)
	}
	for i := range versions {
		versions[i].Component = comp.Component.String
		versions[i].PurlName = comp.PurlName
	}
	m.s.Debugf("Found %v versions for %v.", len(versions), purlID)
	return versions, nil
}
