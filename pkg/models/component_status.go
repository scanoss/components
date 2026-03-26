// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2026 SCANOSS.COM
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

	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	purlhelper "github.com/scanoss/go-purl-helper/pkg"
	"go.uber.org/zap"
)

type ComponentStatusModel struct {
	ctx context.Context
	s   *zap.SugaredLogger
	q   *database.DBQueryContext
}

// ComponentVersionStatus represents the status information for a specific version
type ComponentVersionStatus struct {
	PurlName                string         `db:"purl_name"`
	Version                 string         `db:"version"`
	IndexedDate             sql.NullString `db:"indexed_date"`
	VersionStatus           sql.NullString `db:"version_status"`
	VersionStatusChangeDate sql.NullString `db:"version_status_change_date"`
}

// ComponentProjectStatus represents the status information for a component (ignoring version)
type ComponentProjectStatus struct {
	PurlName         string         `db:"purl_name"`
	Component        string         `db:"component"`
	FirstIndexedDate sql.NullString `db:"first_indexed_date"`
	LastIndexedDate  sql.NullString `db:"last_indexed_date"`
	Status           sql.NullString `db:"status"`
	StatusChangeDate sql.NullString `db:"status_change_date"`
}

// ComponentFullStatus combines version and project status information
type ComponentFullStatus struct {
	ComponentVersionStatus
	ComponentProjectStatus
}

func NewComponentStatusModel(ctx context.Context, s *zap.SugaredLogger, q *database.DBQueryContext) *ComponentStatusModel {
	return &ComponentStatusModel{ctx: ctx, s: s, q: q}
}

// GetComponentStatusByPurlAndVersion gets status information for a specific component version
func (m *ComponentStatusModel) GetComponentStatusByPurlAndVersion(purlString, version string) (*ComponentVersionStatus, error) {
	if len(purlString) == 0 {
		m.s.Errorf("Please specify a valid Purl String to query")
		return nil, errors.New("please specify a valid Purl String to query")
	}
	purl, err := purlhelper.PurlFromString(purlString)
	if err != nil {
		return nil, err
	}
	purlName, err := purlhelper.PurlNameFromString(purlString)
	if err != nil {
		return nil, err
	}
	var status ComponentVersionStatus
	// Query to get both version and component status
	query := `
	SELECT DISTINCT  au.purl_name, au."version",  au.indexed_date,  au.version_status,  au.version_status_change_date 
	FROM 
	 	all_urls au,
	 	mines m
	WHERE
		au.mine_id = m.id AND au.purl_name = $1 AND m.purl_type = $2 AND au."version" = $3 
	`
	var results []ComponentVersionStatus
	err = m.q.SelectContext(m.ctx, &results, query, purlName, purl.Type, version)
	if err != nil {
		m.s.Errorf("Failed to query component status for %v version %v: %v", purlName, version, err)
		return nil, fmt.Errorf("failed to query component status: %v", err)
	}
	if len(results) == 0 {
		m.s.Warnf("No status found for %v version %v", purlName, version)
		return nil, fmt.Errorf("component version not found")
	}
	status = results[0]
	m.s.Debugf("Found status for %v version %v", purlName, version)
	return &status, nil
}

// GetComponentStatusByPurl gets status information for the latest version of a component
func (m *ComponentStatusModel) GetComponentStatusByPurl(purlString string) (*ComponentProjectStatus, error) {
	if len(purlString) == 0 {
		m.s.Errorf("Please specify a valid Purl String to query")
		return nil, errors.New("please specify a valid Purl String to query")
	}
	purl, err := purlhelper.PurlFromString(purlString)
	if err != nil {
		return nil, err
	}
	purlName, err := purlhelper.PurlNameFromString(purlString)
	if err != nil {
		return nil, err
	}
	var status ComponentProjectStatus
	// Query to get both version and component status for the latest version
	query := `
	SELECT DISTINCT
		p.component,
  		p.first_indexed_date,
 		p.last_indexed_date ,
 		p.status,
 		p.status_change_date
	FROM
 		projects p,
 		mines m
	WHERE
 		p.mine_id = m.id
 		AND p.purl_name = $1
 		AND m.purl_type = $2;
	`
	var results []ComponentProjectStatus
	err = m.q.SelectContext(m.ctx, &results, query, purlName, purl.Type)
	if err != nil {
		m.s.Errorf("Failed to query component status for %v: %v", purlName, err)
		return nil, fmt.Errorf("failed to query component status: %v", err)
	}
	if len(results) == 0 {
		m.s.Warnf("No status found for %v", purlName)
		return nil, fmt.Errorf("component not found")
	}
	status = results[0]
	m.s.Debugf("Found status for %v", purlName)
	return &status, nil
}

// GetProjectStatusByPurl gets only the project-level status (no version information)
func (m *ComponentStatusModel) GetProjectStatusByPurl(purlString string) (*ComponentProjectStatus, error) {
	if len(purlString) == 0 {
		m.s.Errorf("Please specify a valid Purl String to query")
		return nil, errors.New("please specify a valid Purl String to query")
	}
	purl, err := purlhelper.PurlFromString(purlString)
	if err != nil {
		return nil, err
	}
	purlName, err := purlhelper.PurlNameFromString(purlString)
	if err != nil {
		return nil, err
	}
	var status ComponentProjectStatus
	query := `
		SELECT DISTINCT
			p.purl_name,
			p.component,
			p.first_indexed_date,
			p.last_indexed_date,
			p.status,
			p.status_change_date
		FROM projects p
		JOIN mines m ON p.mine_id = m.id
		WHERE p.purl_name = $1
			AND m.purl_type = $2
		LIMIT 1
	`
	var results []ComponentProjectStatus
	err = m.q.SelectContext(m.ctx, &results, query, purlName, purl.Type)
	if err != nil {
		m.s.Errorf("Failed to query project status for %v: %v", purlName, err)
		return nil, fmt.Errorf("failed to query project status: %v", err)
	}
	if len(results) == 0 {
		m.s.Warnf("No project status found for %v", purlName)
		return nil, fmt.Errorf("component not found")
	}
	status = results[0]
	m.s.Debugf("Found project status for %v", purlName)
	return &status, nil
}
