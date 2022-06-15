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
	"errors"
	"github.com/jmoiron/sqlx"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/componentsv2"
	zlog "scanoss.com/components/pkg/logger"
	"scanoss.com/components/pkg/usecase"
)

type componentServer struct {
	pb.ComponentsServer
	db *sqlx.DB
}

func NewComponentServer(db *sqlx.DB) pb.ComponentsServer {
	return &componentServer{db: db}
}

// Echo sends back the same message received
func (d componentServer) Echo(ctx context.Context, request *common.EchoRequest) (*common.EchoResponse, error) {
	zlog.S.Infof("Received (%v): %v", ctx, request.GetMessage())
	return &common.EchoResponse{Message: request.GetMessage()}, nil
}

// Search and retrieves a list of components
//TODO: Close db connection after exit this method!
func (d componentServer) SearchComponents(ctx context.Context, request *pb.CompSearchRequest) (*pb.CompSearchResponse, error) {
	zlog.S.Infof("Processing component request: %v", request)

	dtoRequest, err := convertSearchComponentInput(request) // Convert to internal DTO for processing
	if err != nil {
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problem parsing component input data"}
		return &pb.CompSearchResponse{Status: &statusResp}, errors.New("problem parsing component input data")
	}

	// Search the KB for information about the components
	compUc := usecase.NewComponents(ctx, d.db)
	dtoComponents, err := compUc.SearchComponents(dtoRequest)
	if err != nil {
		zlog.S.Errorf("Failed to get components: %v", err)
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problems encountered extracting components data"}
		return &pb.CompSearchResponse{Status: &statusResp}, nil
	}
	zlog.S.Debugf("Parsed Components: %+v", dtoComponents)
	componentsResponse, err := convertSearchComponentOutput(dtoComponents) // Convert the internal data into a response object
	if err != nil {
		zlog.S.Errorf("Failed to convert parsed components: %v", err)
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problems encountered extracting components data"}
		return &pb.CompSearchResponse{Status: &statusResp}, nil
	}
	// Set the status and respond with the data
	statusResp := common.StatusResponse{Status: common.StatusCode_SUCCESS, Message: "Success"}
	return &pb.CompSearchResponse{Components: componentsResponse.Components, Status: &statusResp}, nil
}

func (d componentServer) GetComponentVersions(ctx context.Context, request *pb.CompVersionRequest) (*pb.CompVersionResponse, error) {
	//Verify the input request
	if len(request.Purl) == 0 {
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "There is no purl to retrieve component"}
		return &pb.CompVersionResponse{Status: &statusResp}, errors.New("There is no purl to retrieve component")
	}

	//Convert the request to internal DTO
	dtoRequest, err := convertCompVersionsInput(request)
	if err != nil {
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problem parsing component version input data"}
		return &pb.CompVersionResponse{Status: &statusResp}, errors.New("problem parsing component version input data")
	}

	// Creates the use case
	compUc := usecase.NewComponents(ctx, d.db)
	dtoOutput, err := compUc.GetComponentVersions(dtoRequest)

	if err != nil {
		zlog.S.Errorf("Failed to get components: %v", err)
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problems encountered extracting components data"}
		return &pb.CompVersionResponse{Status: &statusResp}, nil
	}

	reqResponse, err := convertCompVersionsOutput(dtoOutput)
	if err != nil {
		zlog.S.Errorf("Failed to convert parsed components: %v", err)
		statusResp := common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problems encountered extracting components data"}
		return &pb.CompVersionResponse{Status: &statusResp}, nil
	}
	statusResp := common.StatusResponse{Status: common.StatusCode_SUCCESS, Message: "Success"}
	return &pb.CompVersionResponse{Component: reqResponse.Component, Status: &statusResp}, nil
}
