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

package usecase

import (
	"context"
	"github.com/jmoiron/sqlx"
	"scanoss.com/components/pkg/dtos"
	zlog "scanoss.com/components/pkg/logger"
	"scanoss.com/components/pkg/models"
)

type ComponentUseCase struct {
	ctx        context.Context
	conn       *sqlx.Conn
	components *models.ComponentModel
}

func NewComponents(ctx context.Context, conn *sqlx.Conn) *ComponentUseCase {
	return &ComponentUseCase{ctx: ctx, conn: conn,
		components: models.NewComponentModel(ctx, conn),
	}
}

func (c ComponentUseCase) GetComponents(request dtos.ComponentSearchInput) (dtos.ComponentsSearchResults, error) {

	searchResults, err := c.components.GetComponentsByNameType(request.Component, request.Package, -1, -1)
	if err != nil {
		zlog.S.Errorf("Problem encountered searching for components: %v - %v.", request.Component, request.Package)
	}

	var componentsSearchResults []dtos.ComponentSearchResult

	for _, component := range searchResults {
		var componentSearchResult dtos.ComponentSearchResult
		componentSearchResult.Component = component.Component
		componentSearchResult.Purl = "pkg:" + component.PurlType + "/" + component.PurlName
		componentSearchResult.Url = component.Url

		componentsSearchResults = append(componentsSearchResults, componentSearchResult)
	}

	return dtos.ComponentsSearchResults{Components: componentsSearchResults}, nil
}
