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

// Package service implements the gRPC service endpoints
package service

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/componentsv2"
	myconfig "scanoss.com/components/pkg/config"
	se "scanoss.com/components/pkg/errors"
	"scanoss.com/components/pkg/usecase"
	"time"
)

type componentServer struct {
	pb.ComponentsServer
	db     *sqlx.DB
	config *myconfig.ServerConfig
}

func NewComponentServer(db *sqlx.DB, config *myconfig.ServerConfig) pb.ComponentsServer {
	setupMetrics()
	return &componentServer{db: db, config: config}
}

// Echo sends back the same message received
func (d componentServer) Echo(ctx context.Context, request *common.EchoRequest) (*common.EchoResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()
	s.Infof("Received (%v): %v", ctx, request.GetMessage())
	return &common.EchoResponse{Message: request.GetMessage()}, nil
}

// SearchComponents and retrieves a list of components
func (d componentServer) SearchComponents(ctx context.Context, request *pb.CompSearchRequest) (*pb.CompSearchResponse, error) {
	requestStartTime := time.Now() // Capture the scan start time
	s := ctxzap.Extract(ctx).Sugar()
	s.Info("Processing component name request...")
	if len(request.Search) == 0 && len(request.Component) == 0 && len(request.Vendor) == 0 {
		return &pb.CompSearchResponse{Status: se.HandleServiceError(ctx, s, se.NewBadRequestError("No data supplied", nil))}, nil
	}
	dtoRequest, err := convertSearchComponentInput(s, request) // Convert to internal DTO for processing
	if err != nil {
		return &pb.CompSearchResponse{Status: se.HandleServiceError(ctx, s, err)}, nil
	}

	// Search the KB for information about the components
	compUc := usecase.NewComponents(ctx, s, d.db, database.NewDBSelectContext(s, d.db, nil, d.config.Database.Trace))
	dtoComponents, err := compUc.SearchComponents(dtoRequest)
	if err != nil {
		return &pb.CompSearchResponse{Status: se.HandleServiceError(ctx, s, err)}, nil
	}
	s.Debugf("Parsed Components: %+v", dtoComponents)
	componentsResponse, err := convertSearchComponentOutput(s, dtoComponents) // Convert the internal data into a response object
	if err != nil {
		s.Errorf("Failed to convert parsed components: %v", err)
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problems encountered extracting components data"}
		return &pb.CompSearchResponse{Status: &statusResp}, nil
	}
	telemetryCompNameRequestTime(ctx, d.config, requestStartTime) // Record the request processing time
	// Set the status and respond with the data
	statusResp := common.StatusResponse{Status: common.StatusCode_SUCCESS, Message: "Success"}
	return &pb.CompSearchResponse{Components: componentsResponse.Components, Status: &statusResp}, nil
}

func (d componentServer) GetComponentVersions(ctx context.Context, request *pb.CompVersionRequest) (*pb.CompVersionResponse, error) {

	requestStartTime := time.Now() // Capture the scan start time
	s := ctxzap.Extract(ctx).Sugar()
	s.Info("Processing component versions request...")
	//Verify the input request
	if len(request.Purl) == 0 {
		return &pb.CompVersionResponse{Status: se.HandleServiceError(ctx, s, se.NewBadRequestError("No purl supplied", nil))}, nil
	}
	//Convert the request to internal DTO
	dtoRequest, err := convertCompVersionsInput(s, request)
	if err != nil {
		return &pb.CompVersionResponse{Status: se.HandleServiceError(ctx, s, err)}, nil
	}
	// Creates the use case
	compUc := usecase.NewComponents(ctx, s, d.db, database.NewDBSelectContext(s, d.db, nil, d.config.Database.Trace))
	dtoOutput, err := compUc.GetComponentVersions(dtoRequest)
	if err != nil {
		return &pb.CompVersionResponse{Status: se.HandleServiceError(ctx, s, err)}, nil
	}

	reqResponse, err := convertCompVersionsOutput(s, dtoOutput)
	if err != nil {
		s.Errorf("Failed to convert parsed components: %v", err)
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problems encountered extracting components data"}
		return &pb.CompVersionResponse{Status: &statusResp}, nil
	}
	telemetryCompVersionRequestTime(ctx, d.config, requestStartTime)
	// Set the status and respond with the data
	statusResp := common.StatusResponse{Status: common.StatusCode_SUCCESS, Message: "Success"}
	return &pb.CompVersionResponse{Component: reqResponse.Component, Status: &statusResp}, nil
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
