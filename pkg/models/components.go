// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2022 SCANOSS.COM
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
	"github.com/jmoiron/sqlx"
	zlog "scanoss.com/components/pkg/logger"
)

var DEFAULT_PURL_TYPE = "github"
var DEFAULT_MAX_VERSION_LIMIT = 50
var DEFAULT_MAX_COMPONENT_LIMIT = 50

type ComponentModel struct {
	ctx context.Context
	db  *sqlx.DB
}

type Component struct {
	Component string `db:"component"`
	PurlType  string `db:"purl_type"`
	PurlName  string `db:"purl_name"`
	Url       string `db:"-"`
}

func NewComponentModel(ctx context.Context, db *sqlx.DB) *ComponentModel {
	return &ComponentModel{ctx: ctx, db: db}
}

func (m *ComponentModel) GetComponents(search, purlType string, limit, offset int) ([]Component, error) {
	zlog.S.Infof("search parameter: %v", search)
	if len(search) == 0 {
		zlog.S.Error("Please specify a valid Component Name to query")
		return nil, errors.New("please specify a valid component Name to query")
	}

	if limit > DEFAULT_MAX_COMPONENT_LIMIT || limit <= 0 {
		limit = DEFAULT_MAX_COMPONENT_LIMIT
	}

	if offset < 0 {
		offset = 0
	}

	if len(purlType) == 0 {
		purlType = DEFAULT_PURL_TYPE
	}

	limitPerQuery := limit / 6
	if limitPerQuery <= 0 {
		limitPerQuery = 1
	}

	queryJobs := []QueryJob{
		{
			query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.component = $1" +
				" AND m.purl_type = $2" +
				" ORDER BY git_created_at NULLS LAST , git_forks DESC, git_watchers DESC" +
				" LIMIT $3 OFFSET $4;",
			args: []any{search, purlType, limitPerQuery, offset},
		},
		{
			query: "SELECT p.component, p.purl_name, m.purl_type FROM projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.vendor = $1" +
				" AND m.purl_type = $2" +
				" ORDER BY git_created_at NULLS LAST, git_forks DESC, git_watchers DESC" +
				" LIMIT $3 OFFSET $4;",
			args: []any{search, purlType, limitPerQuery, offset},
		},
		{
			query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.purl_name like $1" +
				" AND m.purl_type = $2" +
				" ORDER BY git_created_at NULLS LAST, git_forks DESC, git_watchers DESC" +
				" LIMIT $3 OFFSET $4",
			args: []any{"%" + search + "%" + search + "%", purlType, limitPerQuery, offset},
		},
		{
			query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.purl_name like $1" +
				" AND p.purl_name NOT LIKE $2" +
				" AND m.purl_type = $3" +
				" ORDER BY git_created_at NULLS LAST, git_forks DESC, git_watchers DESC" +
				" LIMIT $4 OFFSET $5",
			args: []any{"%" + search + "%", "%" + search + "%" + search + "%", purlType, limitPerQuery, offset},
		},
		{
			query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.purl_name like $1" +
				" AND m.purl_type = $2" +
				" ORDER BY git_created_at NULLS LAST, git_forks DESC, git_watchers DESC" +
				" LIMIT $3 OFFSET $4",
			args: []any{search + "%", purlType, limitPerQuery, offset},
		},
		{
			query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.purl_name like $1" +
				" AND m.purl_type = $2" +
				" ORDER BY git_created_at NULLS LAST, git_forks DESC, git_watchers DESC" +
				" LIMIT $3 OFFSET $4",
			args: []any{"%" + search, purlType, limitPerQuery, offset},
		},
	}

	allComponents, _ := RunQueriesInParallel[Component](m.db, m.ctx, queryJobs)
	// Valoration
	allComponents = RemoveDuplicated[Component](allComponents)
	return allComponents, nil
}

func (m *ComponentModel) GetComponentsByNameType(compName, purlType string, limit, offset int) ([]Component, error) {

	var allComponents []Component
	if len(compName) == 0 {
		zlog.S.Error("Please specify a valid Component Name to query")
		return allComponents, errors.New("please specify a valid component Name to query")
	}

	if limit > DEFAULT_MAX_COMPONENT_LIMIT || limit <= 0 {
		limit = DEFAULT_MAX_COMPONENT_LIMIT
	}

	if offset < 0 {
		offset = 0
	}

	if len(purlType) == 0 {
		purlType = DEFAULT_PURL_TYPE
	}

	con, err := m.db.Connx(m.ctx)
	defer CloseConn(con)

	if err != nil {
		zlog.S.Errorf("Failed to get a database connection from the pool: %v", err)
		return allComponents, err
	}

	err = con.SelectContext(m.ctx, &allComponents,
		"SELECT component, purl_name, m.purl_type FROM projects p "+
			" LEFT JOIN mines m ON p.mine_id = m.id"+
			" WHERE p.component LIKE $1"+
			" AND m.purl_type = $2"+
			" LIMIT $3 OFFSET $4",
		compName, purlType, limit, offset)

	if err != nil {
		zlog.S.Errorf("Error: Failed to query projects table for %v, %v: %v", compName, purlType, err)
		return allComponents, fmt.Errorf("failed to query the projects table: %v", err)
	}
	allComponents = RemoveDuplicated[Component](allComponents)
	return allComponents, nil
}

func (m *ComponentModel) GetComponentsByVendorType(vendorName, purlType string, limit, offset int) ([]Component, error) {
	return []Component{}, nil
}

func (m *ComponentModel) GetComponentsByNameVendorType(compName, vendor, purlType string, limit, offset int) ([]Component, error) {
	return []Component{}, nil
}
