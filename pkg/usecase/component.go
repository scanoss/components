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
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	purlhelper "github.com/scanoss/go-purl-helper/pkg"
	"go.uber.org/zap"
	"scanoss.com/components/pkg/dtos"
	"scanoss.com/components/pkg/models"
)

type ComponentUseCase struct {
	ctx        context.Context
	s          *zap.SugaredLogger
	q          *database.DBQueryContext
	components *models.ComponentModel
	allUrl     *models.AllUrlsModel
}

func NewComponents(ctx context.Context, s *zap.SugaredLogger, db *sqlx.DB, q *database.DBQueryContext) *ComponentUseCase {
	return &ComponentUseCase{ctx: ctx, s: s, q: q,
		components: models.NewComponentModel(ctx, s, q, database.GetLikeOperator(db)),
		allUrl:     models.NewAllUrlModel(ctx, s, q),
	}
}

func (c ComponentUseCase) SearchComponents(request dtos.ComponentSearchInput) (dtos.ComponentsSearchOutput, error) {
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
		c.s.Errorf("Problem encountered searching for components: %v - %v.", request.Component, request.Package)
	}
	for i := range searchResults {
		searchResults[i].Url, _ = purlhelper.ProjectUrl(searchResults[i].PurlName, searchResults[i].PurlType)
	}
	var componentsSearchResults []dtos.ComponentSearchOutput

	for _, component := range searchResults {
		var componentSearchResult dtos.ComponentSearchOutput
		componentSearchResult.Component = component.Component
		componentSearchResult.Purl = "pkg:" + component.PurlType + "/" + component.PurlName
		componentSearchResult.Url = component.Url

		componentsSearchResults = append(componentsSearchResults, componentSearchResult)
	}
	return dtos.ComponentsSearchOutput{Components: componentsSearchResults}, nil
}

func (c ComponentUseCase) GetComponentVersions(request dtos.ComponentVersionsInput) (dtos.ComponentVersionsOutput, error) {

	if len(request.Purl) == 0 {
		c.s.Errorf("The request does not contains purl to retrieve component versions")
		return dtos.ComponentVersionsOutput{}, errors.New("the request does not contains purl to retrieve component versions")
	}

	allUrls, err := c.allUrl.GetUrlsByPurlString(request.Purl, request.Limit)
	if err != nil {
		c.s.Errorf("Problem encountered gettings URLs versions for: %v - %v.", request.Purl, err)
		return dtos.ComponentVersionsOutput{}, err
	}
	purl, err := purlhelper.PurlFromString(request.Purl)
	if err != nil {
		c.s.Warnf("Problem encountered generating output component versions for: %v - %v.", request.Purl, err)
	}

	purlName := purl.Name
	if purl.Type == "github" {
		purlName = fmt.Sprintf("%s/%s", purl.Namespace, purl.Name)
	}

	projectURL, err := purlhelper.ProjectUrl(purlName, purl.Type)
	if err != nil {
		c.s.Warnf("Problem generating the project URL: %v - %v.", request.Purl, err)
	}

	var output dtos.ComponentOutput
	output.Purl = request.Purl
	if len(allUrls) > 0 {
		output.Url = projectURL
		output.Component = allUrls[0].Component
		output.Versions = []dtos.ComponentVersion{}
		for _, u := range allUrls {
			var version dtos.ComponentVersion
			var license dtos.ComponentLicense
			if len(u.Version) == 0 {
				c.s.Infof("Empty version string supplied for: %+v. Skipping", u)
				continue
			}
			version.Version = u.Version
			if len(u.License) == 0 {
				c.s.Infof("Empty license string supplied for: %+v. Skipping", u)
				version.Licenses = []dtos.ComponentLicense{}
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
	return dtos.ComponentVersionsOutput{Component: output}, nil
}
