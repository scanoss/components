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

// Package usecase contains the business logic for the components API.
package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	cmpHelper "github.com/scanoss/go-component-helper/componenthelper"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	purlhelper "github.com/scanoss/go-purl-helper/pkg"
	"go.uber.org/zap"
	"scanoss.com/components/pkg/config"
	"scanoss.com/components/pkg/dtos"
	se "scanoss.com/components/pkg/errors"
	"scanoss.com/components/pkg/models"
)

type ComponentUseCase struct {
	ctx             context.Context
	s               *zap.SugaredLogger
	q               *database.DBQueryContext
	components      *models.ComponentModel
	allURL          *models.AllURLsModel
	componentStatus *models.ComponentStatusModel
	db              *sqlx.DB
	statusMapper    *config.StatusMapper
}

func NewComponents(ctx context.Context, s *zap.SugaredLogger, db *sqlx.DB, q *database.DBQueryContext, statusMapper *config.StatusMapper) *ComponentUseCase {
	return &ComponentUseCase{ctx: ctx, s: s, q: q,
		components:      models.NewComponentModel(ctx, s, q, database.GetLikeOperator(db)),
		allURL:          models.NewAllURLModel(ctx, s, q),
		componentStatus: models.NewComponentStatusModel(ctx, s, q),
		db:              db,
		statusMapper:    statusMapper,
	}
}

func (c ComponentUseCase) SearchComponents(request dtos.ComponentSearchInput) (dtos.ComponentsSearchOutput, error) {
	var err error
	var searchResults []models.Component
	switch {
	case len(request.Search) != 0:
		searchResults, err = c.components.GetComponents(request.Search, request.Package, request.Limit, request.Offset)
	case len(request.Component) != 0 && len(request.Vendor) == 0:
		searchResults, err = c.components.GetComponentsByNameType(request.Component, request.Package, request.Limit, request.Offset)
	case len(request.Component) == 0 && len(request.Vendor) != 0:
		searchResults, err = c.components.GetComponentsByVendorType(request.Vendor, request.Package, request.Limit, request.Offset)
	case len(request.Component) != 0 && len(request.Vendor) != 0:
		searchResults, err = c.components.GetComponentsByNameVendorType(request.Component, request.Vendor, request.Package, request.Limit, request.Offset)
	}
	if err != nil {
		c.s.Errorf("Problem encountered searching for components: %v - %v.", request.Component, request.Package)
	}
	for i := range searchResults {
		searchResults[i].URL, _ = purlhelper.ProjectUrl(searchResults[i].PurlName, searchResults[i].PurlType)
	}
	var componentsSearchResults []dtos.ComponentSearchOutput

	for _, component := range searchResults {
		var componentSearchResult dtos.ComponentSearchOutput
		componentSearchResult.Name = component.Component
		componentSearchResult.Component = component.Component // Deprecated. Remove in future versions
		componentSearchResult.Purl = "pkg:" + component.PurlType + "/" + component.PurlName
		componentSearchResult.URL = component.URL
		componentsSearchResults = append(componentsSearchResults, componentSearchResult)
	}
	if len(componentsSearchResults) == 0 {
		return dtos.ComponentsSearchOutput{}, se.NewNotFoundError("No components found matching the search criteria")
	}
	return dtos.ComponentsSearchOutput{Components: componentsSearchResults}, nil
}

func (c ComponentUseCase) GetComponentVersions(request dtos.ComponentVersionsInput) (dtos.ComponentVersionsOutput, error) {
	if len(request.Purl) == 0 {
		c.s.Errorf("The request does not contains purl to retrieve component versions")
		return dtos.ComponentVersionsOutput{}, errors.New("the request does not contains purl to retrieve component versions")
	}
	allUrls, err := c.allURL.GetUrlsByPurlString(request.Purl, request.Limit)
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
		output.Name = allUrls[0].Component
		output.URL = projectURL
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
			version.Date = u.Date.String
			if len(u.License) == 0 {
				c.s.Infof("Empty license string supplied for: %+v. Skipping", u)
				version.Licenses = []dtos.ComponentLicense{}
				output.Versions = append(output.Versions, version)
				continue
			}
			license.Name = u.License
			license.SpdxID = u.LicenseID
			license.IsSpdx = u.IsSpdx
			version.Licenses = append(version.Licenses, license)
			output.Versions = append(output.Versions, version)
		}
	}
	if output.Name == "" || output.Purl == "" {
		return dtos.ComponentVersionsOutput{}, se.NewNotFoundError(fmt.Sprintf("purl: '%v' not found", request.Purl))
	}
	return dtos.ComponentVersionsOutput{Component: output}, nil
}

func (c ComponentUseCase) GetComponentStatus(request dtos.ComponentStatusInput) (dtos.ComponentStatusOutput, error) {
	if len(request.Purl) == 0 {
		c.s.Errorf("The request does not contain purl to retrieve component status")
		return dtos.ComponentStatusOutput{}, se.NewBadRequestError("purl is required", errors.New("purl is required"))
	}
	results := cmpHelper.GetComponentsVersion(cmpHelper.ComponentVersionCfg{
		MaxWorkers: 1,
		Ctx:        c.ctx,
		S:          c.s,
		DB:         c.db,
		Input: []cmpHelper.ComponentDTO{
			{Purl: request.Purl, Requirement: request.Requirement},
		},
	})
	if len(results) > 0 {
		return c.handleComponentStatusResult(request, results[0])
	}
	return dtos.ComponentStatusOutput{}, se.NewBadRequestError("purl is required", errors.New("purl is required"))
}

// handleComponentStatusResult routes the component status result to the appropriate handler based on status code.
func (c ComponentUseCase) handleComponentStatusResult(request dtos.ComponentStatusInput, result cmpHelper.Component) (dtos.ComponentStatusOutput, error) {
	//nolint:exhaustive
	switch result.Status.StatusCode {
	case domain.Success:
		return c.handleSuccessStatus(request, result)
	case domain.VersionNotFound:
		return c.handleVersionNotFound(request, result)
	case domain.InvalidPurl, domain.ComponentNotFound:
		return c.handleErrorStatus(result)
	default:
		return dtos.ComponentStatusOutput{}, se.NewBadRequestError("unknown status code", errors.New("unknown status code"))
	}
}

// handleSuccessStatus handles the case where both component and version are found.
func (c ComponentUseCase) handleSuccessStatus(request dtos.ComponentStatusInput, result cmpHelper.Component) (dtos.ComponentStatusOutput, error) {
	statComponent, errComp := c.componentStatus.GetComponentStatusByPurl(result.Purl)
	if errComp != nil {
		return dtos.ComponentStatusOutput{}, se.NewBadRequestError("error retrieving Component level data", errors.New("error retrieving Component Level Data"))
	}
	output := dtos.ComponentStatusOutput{
		Purl:            request.Purl,
		Name:            statComponent.Component,
		Requirement:     request.Requirement,
		ComponentStatus: c.buildComponentStatusInfo(statComponent),
	}
	// Try to get version-specific status
	statusVersion := c.getVersionStatus(request.Purl, result)
	if statusVersion != nil {
		output.VersionStatus = c.buildVersionStatusOutput(statusVersion)
	}
	return output, nil
}

// handleVersionNotFound handles the case where component exists but the version is not found.
func (c ComponentUseCase) handleVersionNotFound(request dtos.ComponentStatusInput, result cmpHelper.Component) (dtos.ComponentStatusOutput, error) {
	statComponent, errComp := c.componentStatus.GetComponentStatusByPurl(result.Purl)
	if errComp != nil {
		return dtos.ComponentStatusOutput{}, se.NewBadRequestError("error retrieving information", errors.New("error retrieving information"))
	}
	return dtos.ComponentStatusOutput{
		Purl:        request.Purl,
		Name:        statComponent.Component,
		Requirement: request.Requirement,
		VersionStatus: &dtos.VersionStatusOutput{
			Version:      request.Requirement,
			ErrorMessage: &result.Status.Message,
			ErrorCode:    &result.Status.StatusCode,
		},
		ComponentStatus: c.buildComponentStatusInfo(statComponent),
	}, nil
}

// handleErrorStatus handles error cases like InvalidPurl or ComponentNotFound.
func (c ComponentUseCase) handleErrorStatus(result cmpHelper.Component) (dtos.ComponentStatusOutput, error) {
	return dtos.ComponentStatusOutput{
		Purl:        result.Purl,
		Name:        "",
		Requirement: result.Requirement,
		ComponentStatus: &dtos.ComponentStatusInfo{
			ErrorMessage: &result.Status.Message,
			ErrorCode:    &result.Status.StatusCode,
		},
	}, nil
}

// buildComponentStatusInfo constructs a ComponentStatusInfo from a ComponentProjectStatus model.
func (c ComponentUseCase) buildComponentStatusInfo(statComponent *models.ComponentProjectStatus) *dtos.ComponentStatusInfo {
	info := &dtos.ComponentStatusInfo{
		Status:           c.statusMapper.MapStatus(statComponent.Status.String),
		RepositoryStatus: statComponent.Status.String,
		FirstIndexedDate: statComponent.FirstIndexedDate.String,
		LastIndexedDate:  statComponent.LastIndexedDate.String,
	}
	if statComponent.StatusChangeDate.String != "" {
		info.StatusChangeDate = statComponent.StatusChangeDate.String
	}
	return info
}

// getVersionStatus retrieves version-specific status information.
func (c ComponentUseCase) getVersionStatus(purl string, result cmpHelper.Component) *models.ComponentVersionStatus {
	var statusVersion *models.ComponentVersionStatus
	var errVersion error
	if len(result.Version) > 0 {
		statusVersion, errVersion = c.componentStatus.GetComponentStatusByPurlAndVersion(purl, result.Version)
	} else if len(result.Requirement) > 0 {
		statusVersion, errVersion = c.componentStatus.GetComponentStatusByPurlAndVersion(purl, result.Requirement)
	}
	if errVersion != nil {
		c.s.Warnf("Problems getting version level status data for: %v - %v", purl, errVersion)
		return nil
	}
	return statusVersion
}

// buildVersionStatusOutput constructs a VersionStatusOutput from a ComponentVersionStatus model.
func (c ComponentUseCase) buildVersionStatusOutput(statusVersion *models.ComponentVersionStatus) *dtos.VersionStatusOutput {
	output := &dtos.VersionStatusOutput{
		Version:          statusVersion.Version,
		Status:           c.statusMapper.MapStatus(statusVersion.VersionStatus.String),
		RepositoryStatus: statusVersion.VersionStatus.String,
		IndexedDate:      statusVersion.IndexedDate.String,
	}
	if statusVersion.VersionStatusChangeDate.String != "" {
		output.StatusChangeDate = statusVersion.VersionStatusChangeDate.String
	}
	return output
}

func (c ComponentUseCase) GetComponentsStatus(request dtos.ComponentsStatusInput) (dtos.ComponentsStatusOutput, error) {
	if len(request.Components) == 0 {
		c.s.Errorf("The request does not contain any components to retrieve status")
		return dtos.ComponentsStatusOutput{}, se.NewBadRequestError("components array is required", errors.New("components array is required"))
	}
	var output dtos.ComponentsStatusOutput
	output.Components = make([]dtos.ComponentStatusOutput, 0, len(request.Components))
	// Process each component request
	for _, componentRequest := range request.Components {
		componentStatus, err := c.GetComponentStatus(componentRequest)
		if err != nil {
			// For batch requests, we continue even if one component fails
			// Add an error entry for this component
			c.s.Warnf("Failed to get status for component: %v - %v", componentRequest.Purl, err)
			errorMsg := err.Error()
			errorStatus := dtos.ComponentStatusOutput{
				Purl:        componentRequest.Purl,
				Name:        "",
				Requirement: componentRequest.Requirement,
				ComponentStatus: &dtos.ComponentStatusInfo{
					ErrorMessage: dtos.StringPtr(errorMsg),
				},
			}
			output.Components = append(output.Components, errorStatus)
		} else {
			output.Components = append(output.Components, componentStatus)
		}
	}
	return output, nil
}
