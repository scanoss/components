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
	"github.com/jmoiron/sqlx"
	"regexp"
	zlog "scanoss.com/components/pkg/logger"
	"strconv"
	"strings"
)

var defaultPurlType = "github"
var defaultMaxVersionLimit = 50
var defaultMaxComponentLimit = 50

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

// preProsessQueryJob Replace the clause #ORDER in the queries (if exist) according to the purlType
// and also adjust the limit per query to the value limit/len(queries)
func preProsessQueryJob(qListIn []QueryJob, purlType string, limit int) ([]QueryJob, error) {

	if len(qListIn) == 0 || limit < 0 {
		return []QueryJob{}, errors.New("Cannot pre process query jobs empty or with limit less than 0")
	}

	qList := make([]QueryJob, len(qListIn))
	copy(qList, qListIn)
	// order by git_created_at NULLS LAST, git_forks DESC NULLS LAST , git_watchers DESC NULLS FIRST
	mapPurlTypeToOrderByClause := map[string]string{
		"github": "ORDER BY git_created_at NULLS LAST , git_forks DESC NULLS LAST, git_watchers DESC NULLS LAST",
		"pypi":   "ORDER BY first_version_date NULLS LAST, versions NULLS LAST",
		"npm":    "ORDER BY first_version_date NULLS LAST, versions NULLS LAST",
		"gem":    "ORDER BY first_version_date NULLS LAST, versions NULLS LAST",
	}

	limitPerQuery := limit / len(qList)
	if limitPerQuery == 0 {
		limitPerQuery = 1
	}

	reLimit, err := regexp.Compile("LIMIT\\s*\\$(\\d)")
	if err != nil {
		return []QueryJob{}, err
	}

	for i, _ := range qList {
		//Adds or remove the ORDER BY clause in SQL query
		qList[i].Query = strings.Replace(qList[i].Query, "#ORDER", mapPurlTypeToOrderByClause[purlType], 1)
		qList[i].Query = strings.TrimRight(qList[i].Query, " ")

		// Extract the arg position for LIMIT statement in the SQL query
		// Then update the value of the limit with limitPerQuery
		positionOfLimitString := reLimit.FindStringSubmatch(qList[i].Query)
		if len(positionOfLimitString) >= 1 {
			positionOfLimit, err := strconv.Atoi(positionOfLimitString[1])
			if err != nil {
				// Cannot update limit value
				zlog.S.Error("Unable to extract the position of the LIMIT argument and update the limit argument in SQL query")
				continue
			}
			qList[i].Args[positionOfLimit-1] = limitPerQuery
		}

	}

	return qList, nil
}

func (m *ComponentModel) GetComponents(search, purlType string, limit, offset int) ([]Component, error) {
	zlog.S.Infof("search parameter: %v", search)
	if len(search) == 0 {
		zlog.S.Error("Please specify a valid Component Name to query")
		return nil, errors.New("please specify a valid component Name to query")
	}

	if limit > defaultMaxComponentLimit || limit <= 0 {
		limit = defaultMaxComponentLimit
	}

	if offset < 0 {
		offset = 0
	}

	if len(purlType) == 0 {
		purlType = defaultPurlType
	}

	queryJobs := []QueryJob{
		{
			Query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.component ILIKE $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4;",
			Args: []any{search, purlType, 1, offset},
		},
		{
			Query: "SELECT p.component, p.purl_name, m.purl_type FROM projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.vendor ILIKE $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4;",
			Args: []any{search, purlType, 1, offset},
		},
		{
			Query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.purl_name ILIKE $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{"%" + search + "%" + search + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.purl_name ILIKE $1" +
				" AND p.purl_name NOT ILIKE $2" +
				" AND m.purl_type = $3" +
				" #ORDER " +
				" LIMIT $4 OFFSET $5",
			Args: []any{"%" + search + "%", "%" + search + "%" + search + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.purl_name ILIKE $1" +
				" AND m.purl_type = $2" +
				" #ORDER " +
				" LIMIT $3 OFFSET $4",
			Args: []any{search + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.purl_name LIKE $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{"%" + search, purlType, 1, offset},
		},
	}

	queryJobs, err := preProsessQueryJob(queryJobs, purlType, limit)
	if err != nil {
		return []Component{}, err
	}

	allComponents, _ := RunQueriesInParallel[Component](m.db, m.ctx, queryJobs)
	allComponents = RemoveDuplicated[Component](allComponents)
	return allComponents, nil
}

func (m *ComponentModel) GetComponentsByNameType(compName, purlType string, limit, offset int) ([]Component, error) {
	if len(compName) == 0 {
		zlog.S.Error("Please specify a valid Component Name to query")
		return []Component{}, errors.New("please specify a valid component Name to query")
	}

	if limit > defaultMaxComponentLimit || limit <= 0 {
		limit = defaultMaxComponentLimit
	}

	if offset < 0 {
		offset = 0
	}

	if len(purlType) == 0 {
		purlType = defaultPurlType
	}

	queryJobs := []QueryJob{
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.component ILIKE $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{compName, purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.component ILIKE $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{"%" + compName + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.component ILIKE $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{compName + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.component ILIKE $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{"%" + compName, purlType, 1, offset},
		},
	}

	queryJobs, err := preProsessQueryJob(queryJobs, purlType, limit)
	if err != nil {
		return []Component{}, err
	}

	allComponents, _ := RunQueriesInParallel[Component](m.db, m.ctx, queryJobs)
	allComponents = RemoveDuplicated[Component](allComponents)
	return allComponents, nil
}

func (m *ComponentModel) GetComponentsByVendorType(vendorName, purlType string, limit, offset int) ([]Component, error) {
	if len(vendorName) == 0 {
		zlog.S.Error("Please specify a valid Component Name to query")
		return []Component{}, errors.New("please specify a valid component Name to query")
	}

	if limit > defaultMaxComponentLimit || limit <= 0 {
		limit = defaultMaxComponentLimit
	}

	if offset < 0 {
		offset = 0
	}

	if len(purlType) == 0 {
		purlType = defaultPurlType
	}

	queryJobs := []QueryJob{
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.vendor = $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{vendorName, purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.vendor ILIKE $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{"%" + vendorName + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.vendor ILIKE $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{vendorName + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.vendor ILIKE $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{"%" + vendorName, purlType, 1, offset},
		},
	}

	queryJobs, err := preProsessQueryJob(queryJobs, purlType, limit)
	if err != nil {
		return []Component{}, err
	}

	allComponents, _ := RunQueriesInParallel[Component](m.db, m.ctx, queryJobs)
	allComponents = RemoveDuplicated[Component](allComponents)
	return allComponents, nil
}

func (m *ComponentModel) GetComponentsByNameVendorType(compName, vendor, purlType string, limit, offset int) ([]Component, error) {

	if len(compName) == 0 || len(vendor) == 0 {
		zlog.S.Error("Please specify a valid Component Name to query")
		return []Component{}, errors.New("please specify a valid component Name to query")
	}

	if limit > defaultMaxComponentLimit || limit <= 0 {
		limit = defaultMaxComponentLimit
	}

	if offset < 0 {
		offset = 0
	}

	if len(purlType) == 0 {
		purlType = defaultPurlType
	}

	queryJobs := []QueryJob{
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.vendor ILIKE $1 AND p.component ILIKE $2" +
				" AND m.purl_type = $3" +
				" #ORDER" +
				" LIMIT $4 OFFSET $5",
			Args: []any{vendor, compName, purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.vendor ILIKE $1 AND p.component ILIKE $2" +
				" AND m.purl_type = $3" +
				" #ORDER" +
				" LIMIT $4 OFFSET $5",
			Args: []any{"%" + vendor + "%", "%" + compName + "%", purlType, 1, offset},
		},
	}

	queryJobs, err := preProsessQueryJob(queryJobs, purlType, limit)
	if err != nil {
		return []Component{}, err
	}

	allComponents, _ := RunQueriesInParallel[Component](m.db, m.ctx, queryJobs)
	allComponents = RemoveDuplicated[Component](allComponents)
	return allComponents, nil
}
