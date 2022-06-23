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
	"errors"
	"github.com/jmoiron/sqlx"
	"scanoss.com/components/pkg/dtos/dtoGetComponentVersion"
	"scanoss.com/components/pkg/dtos/dtoSearchComponent"
	zlog "scanoss.com/components/pkg/logger"
	"scanoss.com/components/pkg/models"
	"scanoss.com/components/pkg/utils"
)

type ComponentUseCase struct {
	ctx        context.Context
	db         *sqlx.DB
	components *models.ComponentModel
	allUrl     *models.AllUrlsModel
}

func NewComponents(ctx context.Context, db *sqlx.DB) *ComponentUseCase {
	return &ComponentUseCase{ctx: ctx, db: db,
		components: models.NewComponentModel(ctx, db),
		allUrl:     models.NewAllUrlModel(ctx, db),
	}
}

func (c ComponentUseCase) SearchComponents(request dtoSearchComponent.ComponentSearchInput) (dtoSearchComponent.ComponentsSearchOutput, error) {
	var err error
	var searchResults []models.Component

	if len(request.Search) != 0 {
		searchResults, err = c.components.GetComponents(request.Search, request.Package, request.Limit, request.Offset)
	} else if len(request.Component) != 0 && len(request.Vendor) == 0 {
		searchResults, err = c.components.GetComponentsByNameType(request.Component, request.Package, request.Limit, request.Offset)
	} else if len(request.Component) == 0 && len(request.Vendor) != 0 {
		searchResults, err = c.components.GetComponentsByVendorType(request.Vendor, request.Package, request.Limit, request.Offset)
	} else if len(request.Component) != 0 && len(request.Vendor) != 0 {
		searchResults, err = c.components.GetComponentsByNameVendorType(request.Component, request.Vendor, request.Package, request.Limit, request.Offset)
	}

	if err != nil {
		zlog.S.Errorf("Problem encountered searching for components: %v - %v.", request.Component, request.Package)
	}

	for i := range searchResults {
		searchResults[i].Url, _ = utils.ProjectUrl(searchResults[i].PurlName, searchResults[i].PurlType)
	}

	var componentsSearchResults []dtoSearchComponent.ComponentSearchOutput

	for _, component := range searchResults {
		var componentSearchResult dtoSearchComponent.ComponentSearchOutput
		componentSearchResult.Component = component.Component
		componentSearchResult.Purl = "pkg:" + component.PurlType + "/" + component.PurlName
		componentSearchResult.Url = component.Url

		componentsSearchResults = append(componentsSearchResults, componentSearchResult)
	}

	return dtoSearchComponent.ComponentsSearchOutput{Components: componentsSearchResults}, nil
}

func (c ComponentUseCase) GetComponentVersions(request dtoGetComponentVersion.ComponentVersionsInput) (dtoGetComponentVersion.ComponentVersionsOutput, error) {

	if len(request.Purl) == 0 {
		zlog.S.Errorf("The request does not contains purl to retrieve component versions")
		return dtoGetComponentVersion.ComponentVersionsOutput{}, errors.New("The request does not contains purl to retrieve component versions")
	}

	allUrls, err := c.allUrl.GetUrlsByPurlString(request.Purl, request.Limit)
	if err != nil {
		zlog.S.Errorf("Problem encountered gettings URLs versions for: %v - %v.", request.Purl, err)
		return dtoGetComponentVersion.ComponentVersionsOutput{}, err
	}

	purl, err := utils.PurlFromString(request.Purl)
	if err != nil {
		zlog.S.Errorf("Problem encountered generating output component versions for: %v - %v.", request.Purl, err)
		return dtoGetComponentVersion.ComponentVersionsOutput{}, err
	}

	projectURL, err := utils.ProjectUrl(request.Purl, purl.Type)
	if err != nil {
		zlog.S.Errorf("Problem generating the project: %v - %v.", request.Purl, err)
		return dtoGetComponentVersion.ComponentVersionsOutput{}, err
	}

	var output dtoGetComponentVersion.ComponentOutput
	output.Purl = request.Purl
	if len(allUrls) > 0 {
		output.Url = projectURL
		output.Component = allUrls[0].Component
		output.Versions = []dtoGetComponentVersion.ComponentVersion{}
		for _, u := range allUrls {
			var version dtoGetComponentVersion.ComponentVersion
			var license dtoGetComponentVersion.ComponentLicense

			if len(u.Version) == 0 {
				zlog.S.Infof("Empty version string supplied for: %+v. Skipping", u)
				continue
			}

			version.Version = u.Version

			if len(u.License) == 0 {
				zlog.S.Infof("Empty license string supplied for: %+v. Skipping", u)
				version.Licenses = []dtoGetComponentVersion.ComponentLicense{}
				output.Versions = append(output.Versions, version)
				continue
			}

			license.Name = u.License
			license.SpdxId = u.LicenseId
			license.IsSpdx = u.IsSpdx
			version.Licenses = append(version.Licenses, license)
			output.Versions = append(output.Versions, version)
		}
	}
	return dtoGetComponentVersion.ComponentVersionsOutput{Component: output}, nil
}
