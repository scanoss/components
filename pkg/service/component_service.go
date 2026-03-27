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

// Package service implements the gRPC service endpoints
package service

import (
	"context"
	"errors"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	gomodels "github.com/scanoss/go-models/pkg/models"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/componentsv2"
	myconfig "scanoss.com/components/pkg/config"
	se "scanoss.com/components/pkg/errors"
	"scanoss.com/components/pkg/usecase"
)

type componentServer struct {
	pb.ComponentsServer
	db             *sqlx.DB
	config         *myconfig.ServerConfig
	dbVersionModel *gomodels.DBVersionModel
}

func NewComponentServer(db *sqlx.DB, config *myconfig.ServerConfig) pb.ComponentsServer {
	setupMetrics()
	return &componentServer{
		db:             db,
		config:         config,
		dbVersionModel: gomodels.NewDBVersionModel(db),
	}
}

// Echo sends back the same message received.
func (d componentServer) Echo(ctx context.Context, request *common.EchoRequest) (*common.EchoResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()
	s.Infof("Received (%v): %v", ctx, request.GetMessage())
	return &common.EchoResponse{Message: request.GetMessage()}, nil
}

// SearchComponents and retrieves a list of components.
func (d componentServer) SearchComponents(ctx context.Context, request *pb.CompSearchRequest) (*pb.CompSearchResponse, error) {
	requestStartTime := time.Now() // Capture the scan start time
	s := ctxzap.Extract(ctx).Sugar()
	s.Info("Processing component name request...")
	if len(request.Search) == 0 && len(request.Component) == 0 && len(request.Vendor) == 0 {
		status := se.HandleServiceError(ctx, s, se.NewBadRequestError("No data supplied", nil))
		status.Db = d.getDBVersion()
		status.Server = &common.StatusResponse_Server{Version: d.config.App.Version}
		return &pb.CompSearchResponse{Status: status}, nil
	}
	dtoRequest, err := convertSearchComponentInput(s, request) // Convert to internal DTO for processing
	if err != nil {
		status := se.HandleServiceError(ctx, s, err)
		status.Db = d.getDBVersion()
		status.Server = &common.StatusResponse_Server{Version: d.config.App.Version}
		return &pb.CompSearchResponse{Status: status}, nil
	}

	// Search the KB for information about the components
	compUc := usecase.NewComponents(ctx, s, d.db, database.NewDBSelectContext(s, d.db, nil, d.config.Database.Trace), d.config.GetStatusMapper())
	dtoComponents, err := compUc.SearchComponents(dtoRequest)
	if err != nil {
		status := se.HandleServiceError(ctx, s, err)
		status.Db = d.getDBVersion()
		status.Server = &common.StatusResponse_Server{Version: d.config.App.Version}
		return &pb.CompSearchResponse{Status: status}, nil
	}
	s.Debugf("Parsed Components: %+v", dtoComponents)
	componentsResponse, err := convertSearchComponentOutput(s, dtoComponents) // Convert the internal data into a response object
	if err != nil {
		s.Errorf("Failed to convert parsed components: %v", err)
		return &pb.CompSearchResponse{Status: &common.StatusResponse{
			Status:  common.StatusCode_FAILED,
			Message: "Problems encountered extracting components data",
			Db:      d.getDBVersion(),
			Server:  &common.StatusResponse_Server{Version: d.config.App.Version},
		}}, nil
	}
	telemetryCompNameRequestTime(ctx, d.config, requestStartTime) // Record the request processing time
	// Set the status and respond with the data
	return &pb.CompSearchResponse{Components: componentsResponse.Components, Status: &common.StatusResponse{
		Status:  common.StatusCode_SUCCESS,
		Message: "Success",
		Db:      d.getDBVersion(),
		Server:  &common.StatusResponse_Server{Version: d.config.App.Version},
	}}, nil
}

func (d componentServer) GetComponentVersions(ctx context.Context, request *pb.CompVersionRequest) (*pb.CompVersionResponse, error) {
	requestStartTime := time.Now() // Capture the scan start time
	s := ctxzap.Extract(ctx).Sugar()
	s.Info("Processing component versions request...")
	// Verify the input request
	if len(request.Purl) == 0 {
		status := se.HandleServiceError(ctx, s, se.NewBadRequestError("No purl supplied", nil))
		status.Db = d.getDBVersion()
		status.Server = &common.StatusResponse_Server{Version: d.config.App.Version}
		return &pb.CompVersionResponse{Status: status}, nil
	}
	// Convert the request to internal DTO
	dtoRequest, err := convertCompVersionsInput(s, request)
	if err != nil {
		status := se.HandleServiceError(ctx, s, err)
		status.Db = d.getDBVersion()
		status.Server = &common.StatusResponse_Server{Version: d.config.App.Version}
		return &pb.CompVersionResponse{Status: status}, nil
	}
	// Creates the use case
	compUc := usecase.NewComponents(ctx, s, d.db, database.NewDBSelectContext(s, d.db, nil, d.config.Database.Trace), d.config.GetStatusMapper())
	dtoOutput, err := compUc.GetComponentVersions(dtoRequest)
	if err != nil {
		status := se.HandleServiceError(ctx, s, err)
		status.Db = d.getDBVersion()
		status.Server = &common.StatusResponse_Server{Version: d.config.App.Version}
		return &pb.CompVersionResponse{Status: status}, nil
	}

	reqResponse, err := convertCompVersionsOutput(s, dtoOutput)
	if err != nil {
		s.Errorf("Failed to convert parsed components: %v", err)
		return &pb.CompVersionResponse{Status: &common.StatusResponse{
			Status:  common.StatusCode_FAILED,
			Message: "Problems encountered extracting components data",
			Db:      d.getDBVersion(),
			Server:  &common.StatusResponse_Server{Version: d.config.App.Version},
		}}, nil
	}
	telemetryCompVersionRequestTime(ctx, d.config, requestStartTime)
	// Set the status and respond with the data
	return &pb.CompVersionResponse{Component: reqResponse.Component, Status: &common.StatusResponse{
		Status:  common.StatusCode_SUCCESS,
		Message: "Success",
		Db:      d.getDBVersion(),
		Server:  &common.StatusResponse_Server{Version: d.config.App.Version},
	}}, nil
}

// telemetryCompNameRequestTime records the name request time to telemetry.
func telemetryCompNameRequestTime(ctx context.Context, config *myconfig.ServerConfig, requestStartTime time.Time) {
	if config.Telemetry.Enabled {
		elapsedTime := time.Since(requestStartTime).Milliseconds() // Time taken to run the component name request
		oltpMetrics.compNameHistogram.Record(ctx, elapsedTime)     // Record dep request time
	}
}

// telemetryCompNameRequestTime records the versions request time to telemetry.
func telemetryCompVersionRequestTime(ctx context.Context, config *myconfig.ServerConfig, requestStartTime time.Time) {
	if config.Telemetry.Enabled {
		elapsedTime := time.Since(requestStartTime).Milliseconds() // Time taken to run the component version request
		oltpMetrics.compVersionHistogram.Record(ctx, elapsedTime)  // Record dep request time
	}
}

// GetComponentStatus retrieves status information for a specific component.
func (d componentServer) GetComponentStatus(ctx context.Context, request *common.ComponentRequest) (*pb.ComponentStatusResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()
	s.Info("Processing component status request...")
	// Verify the input request
	if len(request.Purl) == 0 {
		s.Error("No purl supplied")
		return &pb.ComponentStatusResponse{}, se.NewBadRequestError("No purl supplied", nil)
	}
	// Convert the request to internal DTO
	dtoRequest, err := convertComponentStatusInput(s, request)
	if err != nil {
		s.Errorf("Failed to convert component status input: %v", err)
		return &pb.ComponentStatusResponse{}, err
	}
	// Create the use case
	compUc := usecase.NewComponents(ctx, s, d.db, database.NewDBSelectContext(s, d.db, nil, d.config.Database.Trace), d.config.GetStatusMapper())
	dtoOutput, err := compUc.GetComponentStatus(dtoRequest)
	if err != nil {
		s.Errorf("Failed to get component status: %v", err)
		return &pb.ComponentStatusResponse{}, err
	}
	// Convert the output to protobuf
	statusResponse := convertComponentStatusOutput(dtoOutput)
	return statusResponse, nil
}

// GetComponentsStatus retrieves status information for multiple components.
func (d componentServer) GetComponentsStatus(ctx context.Context, request *common.ComponentsRequest) (*pb.ComponentsStatusResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()
	s.Info("Processing components status request...")
	// Verify the input request
	if len(request.Components) == 0 {
		status := se.HandleServiceError(ctx, s, se.NewBadRequestError("No components supplied", nil))
		status.Db = d.getDBVersion()
		status.Server = &common.StatusResponse_Server{Version: d.config.App.Version}
		return &pb.ComponentsStatusResponse{Status: status}, nil
	}
	// Convert the request to internal DTO
	dtoRequest, err := convertComponentsStatusInput(s, request)
	if err != nil {
		status := se.HandleServiceError(ctx, s, err)
		status.Db = d.getDBVersion()
		status.Server = &common.StatusResponse_Server{Version: d.config.App.Version}
		return &pb.ComponentsStatusResponse{Status: status}, nil
	}
	// Create the use case
	compUc := usecase.NewComponents(ctx, s, d.db, database.NewDBSelectContext(s, d.db, nil, d.config.Database.Trace), d.config.GetStatusMapper())
	dtoOutput, err := compUc.GetComponentsStatus(dtoRequest)
	if err != nil {
		status := se.HandleServiceError(ctx, s, err)
		status.Db = d.getDBVersion()
		status.Server = &common.StatusResponse_Server{Version: d.config.App.Version}
		return &pb.ComponentsStatusResponse{Status: status}, nil
	}
	// Convert the output to protobuf
	statusResponse := convertComponentsStatusOutput(dtoOutput)
	// Set the status and respond with the data
	return &pb.ComponentsStatusResponse{
		Components: statusResponse.Components,
		Status: &common.StatusResponse{
			Status:  common.StatusCode_SUCCESS,
			Message: "Success",
			Db:      d.getDBVersion(),
			Server:  &common.StatusResponse_Server{Version: d.config.App.Version},
		},
	}, nil
}

// getDBVersion fetches the database version from the db_version table.
// Returns nil if the table doesn't exist or query fails (backward compatibility).
func (d componentServer) getDBVersion() *common.StatusResponse_DB {
	dbVersion, err := d.dbVersionModel.GetCurrentVersion(context.Background())
	if err != nil {
		if !errors.Is(err, gomodels.ErrTableNotFound) {
			s := ctxzap.Extract(context.Background()).Sugar()
			s.Errorf("Failed to get db version: %v", err)
		}
		return nil
	}
	if len(dbVersion.SchemaVersion) == 0 {
		return nil
	}
	return &common.StatusResponse_DB{
		SchemaVersion: dbVersion.SchemaVersion,
		CreatedAt:     dbVersion.CreatedAt,
	}
}
