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
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	"go.uber.org/zap"
	"strings"
)

var defaultPurlType = "github"
var defaultMaxVersionLimit = 50
var defaultMaxComponentLimit = 50
var defaultLikeValue = "LIKE"

type ComponentModel struct {
	ctx          context.Context
	s            *zap.SugaredLogger
	q            *database.DBQueryContext
	likeOperator string
}

type Component struct {
	Component string `db:"component"`
	PurlType  string `db:"purl_type"`
	PurlName  string `db:"purl_name"`
	Url       string `db:"-"`
}

func NewComponentModel(ctx context.Context, s *zap.SugaredLogger, q *database.DBQueryContext, likeOperator string) *ComponentModel {
	if len(likeOperator) == 0 {
		likeOperator = defaultLikeValue
	}
	return &ComponentModel{ctx: ctx, s: s, q: q, likeOperator: likeOperator}
}

// preProcessQueryJob Replace the clause #ORDER in the queries (if exist) according to the purlType
func preProcessQueryJob(qListIn []QueryJob, purlType string) ([]QueryJob, error) {

	if len(qListIn) == 0 {
		return []QueryJob{}, errors.New("cannot pre process query jobs empty or with limit less than 0")
	}

	qList := make([]QueryJob, len(qListIn))
	copy(qList, qListIn)
	// order by git_created_at NULLS LAST, git_forks DESC NULLS LAST , git_watchers DESC NULLS FIRST
	mapPurlTypeToOrderByClause := map[string]string{
		"github": "ORDER BY git_created_at NULLS LAST , git_forks DESC NULLS LAST, git_stars DESC NULLS LAST",
		"pypi":   "ORDER BY first_version_date NULLS LAST, versions NULLS LAST",
		"npm":    "ORDER BY first_version_date NULLS LAST, versions NULLS LAST",
		"gem":    "ORDER BY first_version_date NULLS LAST, versions NULLS LAST",
	}

	for i := range qList {
		//Adds or remove the ORDER BY clause in SQL query
		qList[i].Query = strings.Replace(qList[i].Query, "#ORDER", mapPurlTypeToOrderByClause[purlType], 1)
		qList[i].Query = strings.TrimRight(qList[i].Query, " ")
	}

	return qList, nil
}

func (m *ComponentModel) GetComponents(search, purlType string, limit, offset int) ([]Component, error) {
	m.s.Infof("search parameter: %v", search)
	if len(search) == 0 {
		m.s.Error("Please specify a valid Component Name to query")
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
				" WHERE p.component " + m.likeOperator + " $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4;",
			Args: []any{search, purlType, limit, offset},
		},
		{
			Query: "SELECT p.component, p.purl_name, m.purl_type FROM projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.vendor " + m.likeOperator + " $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4;",
			Args: []any{search, purlType, limit, offset},
		},
		{
			Query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.purl_name " + m.likeOperator + " $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{"%" + search + "%" + search + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.purl_name " + m.likeOperator + " $1" +
				" AND p.purl_name NOT " + m.likeOperator + " $2" +
				" AND m.purl_type = $3" +
				" #ORDER " +
				" LIMIT $4 OFFSET $5",
			Args: []any{"%" + search + "%", "%" + search + "%" + search + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT p.component, p.purl_name, m.purl_type from projects p" +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.purl_name " + m.likeOperator + " $1" +
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

	queryJobs, err := preProcessQueryJob(queryJobs, purlType)
	if err != nil {
		return []Component{}, err
	}

	allComponents, _ := RunQueries[Component](m.q, m.ctx, queryJobs)
	allComponents = RemoveDuplicated[Component](allComponents)

	if limit < len(allComponents) {
		allComponents = allComponents[:limit]
	}

	return allComponents, nil
}

func (m *ComponentModel) GetComponentsByNameType(compName, purlType string, limit, offset int) ([]Component, error) {
	if len(compName) == 0 {
		m.s.Error("Please specify a valid Component Name to query")
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
				" WHERE p.component " + m.likeOperator + " $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{compName, purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.component " + m.likeOperator + " $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{"%" + compName + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.component " + m.likeOperator + " $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{compName + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.component " + m.likeOperator + " $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{"%" + compName, purlType, 1, offset},
		},
	}

	queryJobs, err := preProcessQueryJob(queryJobs, purlType)
	if err != nil {
		return []Component{}, err
	}

	allComponents, _ := RunQueries[Component](m.q, m.ctx, queryJobs)
	allComponents = RemoveDuplicated[Component](allComponents)

	if limit < len(allComponents) {
		allComponents = allComponents[:limit]
	}

	return allComponents, nil
}

func (m *ComponentModel) GetComponentsByVendorType(vendorName, purlType string, limit, offset int) ([]Component, error) {
	if len(vendorName) == 0 {
		m.s.Error("Please specify a valid Component Name to query")
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
				" WHERE p.vendor " + m.likeOperator + " $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{"%" + vendorName + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.vendor " + m.likeOperator + " $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{vendorName + "%", purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.vendor " + m.likeOperator + " $1" +
				" AND m.purl_type = $2" +
				" #ORDER" +
				" LIMIT $3 OFFSET $4",
			Args: []any{"%" + vendorName, purlType, 1, offset},
		},
	}

	queryJobs, err := preProcessQueryJob(queryJobs, purlType)
	if err != nil {
		return []Component{}, err
	}

	allComponents, _ := RunQueries[Component](m.q, m.ctx, queryJobs)
	allComponents = RemoveDuplicated[Component](allComponents)

	if limit < len(allComponents) {
		allComponents = allComponents[:limit]
	}

	return allComponents, nil
}

func (m *ComponentModel) GetComponentsByNameVendorType(compName, vendor, purlType string, limit, offset int) ([]Component, error) {

	if len(compName) == 0 || len(vendor) == 0 {
		m.s.Error("Please specify a valid Component Name to query")
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
				" WHERE p.vendor " + m.likeOperator + " $1 AND p.component " + m.likeOperator + " $2" +
				" AND m.purl_type = $3" +
				" #ORDER" +
				" LIMIT $4 OFFSET $5",
			Args: []any{vendor, compName, purlType, 1, offset},
		},
		{
			Query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.vendor " + m.likeOperator + " $1 AND p.component " + m.likeOperator + " $2" +
				" AND m.purl_type = $3" +
				" #ORDER" +
				" LIMIT $4 OFFSET $5",
			Args: []any{"%" + vendor + "%", "%" + compName + "%", purlType, 1, offset},
		},
	}

	queryJobs, err := preProcessQueryJob(queryJobs, purlType)
	if err != nil {
		return []Component{}, err
	}

	allComponents, _ := RunQueries[Component](m.q, m.ctx, queryJobs)
	allComponents = RemoveDuplicated[Component](allComponents)

	if limit < len(allComponents) {
		allComponents = allComponents[:limit]
	}

	return allComponents, nil
}
