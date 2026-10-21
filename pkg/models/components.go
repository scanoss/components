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
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	"go.uber.org/zap"
)

var defaultPurlType = "github"
var defaultMaxComponentLimit = 50
var defaultLikeValue = "LIKE"

type ComponentModel struct {
	ctx          context.Context
	s            *zap.SugaredLogger
	q            *database.DBQueryContext
	db           *sqlx.DB
	likeOperator string
}

// Component is one search hit. The serving schema stores the whole purl on the
// row, so Purl is read directly and PurlType/PurlName are split back out of it
// for the callers that still build a purl from the pair.
type Component struct {
	Component string `db:"component"`
	Purl      string `db:"purl"`
	PurlType  string `db:"-"`
	PurlName  string `db:"-"`
	URL       string `db:"-"`
}

func NewComponentModel(ctx context.Context, s *zap.SugaredLogger, q *database.DBQueryContext, db *sqlx.DB, likeOperator string) *ComponentModel {
	if len(likeOperator) == 0 {
		likeOperator = defaultLikeValue
	}
	return &ComponentModel{ctx: ctx, s: s, q: q, db: db, likeOperator: likeOperator}
}

// searchFrom is the shared FROM clause. The mining search joined `projects` to
// `mines` to reach the purl type; serving keeps the whole purl on `component`
// and the popularity columns in `component_metrics`, one optional row per
// component, which the ORDER BY ranks on.
const searchFrom = `FROM component c LEFT JOIN component_metrics cm ON cm.purl_id = c.purl_id`

// searchSelect is the projection every search pass shares.
const searchSelect = `SELECT c.purl AS purl, c.component AS component `

// orderByPurlType is the per-ecosystem popularity ordering, with the columns
// qualified for the component_metrics alias. A purl type with no entry gets no
// ORDER BY, as in mining.
var orderByPurlType = map[string]string{
	"github": "ORDER BY cm.git_created_at NULLS LAST, cm.git_forks DESC NULLS LAST, cm.git_stars DESC NULLS LAST",
	"pypi":   "ORDER BY cm.first_version_date NULLS LAST, cm.versions NULLS LAST",
	"npm":    "ORDER BY cm.first_version_date NULLS LAST, cm.versions NULLS LAST",
	"gem":    "ORDER BY cm.first_version_date NULLS LAST, cm.versions NULLS LAST",
}

// typePrefix is the LIKE pattern bounding a purl type: 'pkg:<type>/%'. It
// replaces the `mines.purl_type = ?` filter, since serving has no mines table
// and the type is the purl's first segment.
func typePrefix(purlType string) string { return "pkg:" + purlType + "/%" }

// namePattern folds the fixed type segment into a purl-name pattern, so a match
// on the name is implicitly constrained to the type.
func namePattern(purlType, pattern string) string { return "pkg:" + purlType + "/" + pattern }

// colMatch builds a pass matching one column of `component`, bounded to the purl
// type and ranked by the type's order clause.
func (m *ComponentModel) colMatch(op, column, value, purlType string, limit, offset int) QueryJob {
	return QueryJob{
		Query: m.db.Rebind(fmt.Sprintf("%s%s WHERE c.%s %s ? AND c.purl LIKE ? %s LIMIT ? OFFSET ?",
			searchSelect, searchFrom, column, op, orderByPurlType[purlType])),
		Args: []any{value, typePrefix(purlType), limit, offset},
	}
}

// purlMatch builds a pass matching the purl itself; the type is already folded
// into the pattern.
func (m *ComponentModel) purlMatch(op, pattern, purlType string, limit, offset int) QueryJob {
	return QueryJob{
		Query: m.db.Rebind(fmt.Sprintf("%s%s WHERE c.purl %s ? %s LIMIT ? OFFSET ?",
			searchSelect, searchFrom, op, orderByPurlType[purlType])),
		Args: []any{pattern, limit, offset},
	}
}

// normaliseSearchArgs applies the shared bounds every search entry point uses.
func normaliseSearchArgs(purlType string, limit, offset int) (string, int, int) {
	if limit > defaultMaxComponentLimit || limit <= 0 {
		limit = defaultMaxComponentLimit
	}
	if offset < 0 {
		offset = 0
	}
	if len(purlType) == 0 {
		purlType = defaultPurlType
	}
	return purlType, limit, offset
}

// runSearch executes the passes, drops duplicates and truncates to the limit.
//
// A failing pass is reported rather than discarded. RunQueries abandons the whole
// batch on the first error, and the callers used to drop that error on the floor,
// so a broken query surfaced as "no components found" - indistinguishable from a
// search that genuinely matched nothing.
func (m *ComponentModel) runSearch(queryJobs []QueryJob, limit int) ([]Component, error) {
	allComponents, err := RunQueries[Component](m.q, m.ctx, queryJobs)
	if err != nil {
		m.s.Errorf("Failed to run the component search: %v", err)
		return nil, fmt.Errorf("failed to search components: %w", err)
	}
	allComponents = RemoveDuplicated[Component](allComponents)
	for i := range allComponents {
		allComponents[i].PurlType, allComponents[i].PurlName = splitPurl(allComponents[i].Purl)
	}
	if limit < len(allComponents) {
		allComponents = allComponents[:limit]
	}
	return allComponents, nil
}

// splitPurl breaks "pkg:<type>/<name>" back into its two parts. The mining
// search read them from separate columns; serving stores the assembled purl.
func splitPurl(purl string) (string, string) {
	trimmed := strings.TrimPrefix(purl, "pkg:")
	if idx := strings.Index(trimmed, "/"); idx >= 0 {
		return trimmed[:idx], trimmed[idx+1:]
	}
	return trimmed, ""
}

func (m *ComponentModel) GetComponents(search, purlType string, limit, offset int) ([]Component, error) {
	m.s.Infof("search parameter: %v", search)
	if len(search) == 0 {
		m.s.Error("Please specify a valid Component Name to query")
		return nil, errors.New("please specify a valid component Name to query")
	}
	purlType, limit, offset = normaliseSearchArgs(purlType, limit, offset)
	op := m.likeOperator

	queryJobs := []QueryJob{
		m.colMatch(op, "component", search, purlType, limit, offset),
		m.colMatch(op, "vendor", search, purlType, limit, offset),
		m.purlMatch(op, namePattern(purlType, "%"+search+"%"+search+"%"), purlType, 1, offset),
		// Contains, excluding what the previous pass already covered.
		{
			Query: m.db.Rebind(fmt.Sprintf("%s%s WHERE c.purl %s ? AND c.purl NOT %s ? %s LIMIT ? OFFSET ?",
				searchSelect, searchFrom, op, op, orderByPurlType[purlType])),
			Args: []any{namePattern(purlType, "%"+search+"%"), namePattern(purlType, "%"+search+"%"+search+"%"), 1, offset},
		},
		m.purlMatch(op, namePattern(purlType, search+"%"), purlType, 1, offset),
		// The mining query hardcoded LIKE on this last pass; kept as-is.
		m.purlMatch(defaultLikeValue, namePattern(purlType, "%"+search), purlType, 1, offset),
	}
	return m.runSearch(queryJobs, limit)
}

func (m *ComponentModel) GetComponentsByNameType(compName, purlType string, limit, offset int) ([]Component, error) {
	if len(compName) == 0 {
		m.s.Error("Please specify a valid Component Name to query")
		return []Component{}, errors.New("please specify a valid component Name to query")
	}
	purlType, limit, offset = normaliseSearchArgs(purlType, limit, offset)
	op := m.likeOperator

	return m.runSearch([]QueryJob{
		m.colMatch(op, "component", compName, purlType, 1, offset),
		m.colMatch(op, "component", "%"+compName+"%", purlType, 1, offset),
		m.colMatch(op, "component", compName+"%", purlType, 1, offset),
		m.colMatch(op, "component", "%"+compName, purlType, 1, offset),
	}, limit)
}

func (m *ComponentModel) GetComponentsByVendorType(vendorName, purlType string, limit, offset int) ([]Component, error) {
	if len(vendorName) == 0 {
		m.s.Error("Please specify a valid Component Name to query")
		return []Component{}, errors.New("please specify a valid component Name to query")
	}
	purlType, limit, offset = normaliseSearchArgs(purlType, limit, offset)
	op := m.likeOperator

	return m.runSearch([]QueryJob{
		// The first vendor pass is an exact match, as in mining.
		m.colMatch("=", "vendor", vendorName, purlType, 1, offset),
		m.colMatch(op, "vendor", "%"+vendorName+"%", purlType, 1, offset),
		m.colMatch(op, "vendor", vendorName+"%", purlType, 1, offset),
		m.colMatch(op, "vendor", "%"+vendorName, purlType, 1, offset),
	}, limit)
}

func (m *ComponentModel) GetComponentsByNameVendorType(compName, vendor, purlType string, limit, offset int) ([]Component, error) {
	if len(compName) == 0 || len(vendor) == 0 {
		m.s.Error("Please specify a valid Component Name to query")
		return []Component{}, errors.New("please specify a valid component Name to query")
	}
	purlType, limit, offset = normaliseSearchArgs(purlType, limit, offset)
	op := m.likeOperator

	query := m.db.Rebind(fmt.Sprintf("%s%s WHERE c.vendor %s ? AND c.component %s ? AND c.purl LIKE ? %s LIMIT ? OFFSET ?",
		searchSelect, searchFrom, op, op, orderByPurlType[purlType]))
	return m.runSearch([]QueryJob{
		{Query: query, Args: []any{vendor, compName, typePrefix(purlType), 1, offset}},
		{Query: query, Args: []any{"%" + vendor + "%", "%" + compName + "%", typePrefix(purlType), 1, offset}},
	}, limit)
}
