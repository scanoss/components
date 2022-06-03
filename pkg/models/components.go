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
	"scanoss.com/components/pkg/utils"
)

var MAX_LIMIT = 50

type ComponentModel struct {
	ctx  context.Context
	conn *sqlx.Conn
}

type Component struct {
	Component string `db:"component"`
	PurlType  string `db:"purl_type"`
	PurlName  string `db:"purl_name"`
	Url       string `db:"-"`
}

func NewComponentModel(ctx context.Context, conn *sqlx.Conn) *ComponentModel {
	return &ComponentModel{ctx: ctx, conn: conn}
}

//func (m *componentModel) GetComponentsByGenericSearch(generic string, purlType string, searchParams CompSearchCfg) ([]Component, error) {}
//func (m *componentModel) GetComponentsByVendorSearch(vendor string, purlType string, cfg CompSearchCfg) ([]Component, error) {}
//func (m *componentModel) GetComponentsByCompVendSearch(comp_vendor string, cfg CompSearchCfg) ([]Component, error) {}

func (m *ComponentModel) GetComponentsByName(compName string, limit, offset int) ([]Component, error) {
	var allComponents []Component
	return allComponents, nil
}

func (m *ComponentModel) GetComponentsByNameType(compName, purlType string, limit, offset int) ([]Component, error) {

	if len(compName) == 0 {
		zlog.S.Error("Please specify a valid Component Name to query")
		return nil, errors.New("please specify a valid component Name to query")
	}

	if limit > MAX_LIMIT || limit == -1 {
		limit = MAX_LIMIT
	}

	if (offset < 0) {
		offset = 0
	}

	if len(purlType) == 0 {
		return m.GetComponentsByName(compName, limit, offset)
	}

	var allComponents []Component

	err := m.conn.SelectContext(m.ctx, &allComponents,
		"SELECT component, purl_name, m.purl_type FROM projects p"+
			" LEFT JOIN mines m ON p.mine_id = m.id"+
			" WHERE p.component LIKE $1"+
			" AND m.purl_type = $2"+
			" LIMIT $3 OFFSET $4;",
		compName, purlType, limit, offset)

	if err != nil {
		zlog.S.Errorf("Error: Failed to query projects table for %v, %v: %v", compName, purlType, err)
		return nil, fmt.Errorf("failed to query the projects table: %v", err)
	}

	populateURLIntoComponents(allComponents)

	return allComponents, nil
}

func populateURLIntoComponents(compList []Component) {

	for i, _ := range compList {
		compList[i].Url, _ = utils.ProjectUrl(compList[i].PurlName, compList[i].PurlType)
	}

}
