package service

import (
	"encoding/json"
	"errors"

	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	pb "github.com/scanoss/papi/api/componentsv2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"scanoss.com/components/pkg/dtos"
	se "scanoss.com/components/pkg/errors"
)

// Structure for storing OTEL metrics.
type metricsCounters struct {
	compNameHistogram    metric.Int64Histogram // milliseconds
	compVersionHistogram metric.Int64Histogram // milliseconds
}

var oltpMetrics = metricsCounters{}

// setupMetrics configures all the metrics recorders for the platform.
func setupMetrics() {
	meter := otel.Meter("scanoss.com/components")
	oltpMetrics.compNameHistogram, _ = meter.Int64Histogram("comp.name.req_time", metric.WithDescription("The time taken to run a comp name request (ms)"))
	oltpMetrics.compVersionHistogram, _ = meter.Int64Histogram("comp.versions.req_time", metric.WithDescription("The time taken to run a comp versions request (ms)"))
}

// convertSearchComponentInput converts a gRPC CompSearchRequest into a ComponentSearchInput DTO.
// It marshals the gRPC request to JSON and then unmarshals it into the internal DTO format.
//
// Parameters:
//   - s: Sugared logger for error logging
//   - request: gRPC component search request
//
// Returns:
//   - ComponentSearchInput DTO or BadRequestError if conversion fails
func convertSearchComponentInput(s *zap.SugaredLogger, request *pb.CompSearchRequest) (dtos.ComponentSearchInput, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return dtos.ComponentSearchInput{}, se.NewBadRequestError("Error parsing request data", err)
	}
	dtoRequest, err := dtos.ParseComponentSearchInput(s, data)
	if err != nil {
		return dtos.ComponentSearchInput{}, se.NewBadRequestError("Error parsing request data", err)
	}
	return dtoRequest, nil
}

// convertSearchComponentOutput converts a ComponentsSearchOutput DTO into a gRPC CompSearchResponse.
// It marshals the DTO to JSON and then unmarshals it into the gRPC response format.
//
// Parameters:
//   - s: Sugared logger for error logging
//   - output: ComponentsSearchOutput DTO containing search results
//
// Returns:
//   - gRPC CompSearchResponse or error if conversion fails
func convertSearchComponentOutput(s *zap.SugaredLogger, output dtos.ComponentsSearchOutput) (*pb.CompSearchResponse, error) {
	data, err := json.Marshal(output)
	if err != nil {
		s.Errorf("Problem marshalling component request output: %v", err)
		return &pb.CompSearchResponse{}, errors.New("problem marshalling component output")
	}
	var compResp pb.CompSearchResponse
	err = json.Unmarshal(data, &compResp)
	if err != nil {
		s.Errorf("Problem unmarshalling component request output: %v", err)
		return &pb.CompSearchResponse{}, errors.New("problem unmarshalling component output")
	}
	return &compResp, nil
}

// convertCompVersionsInput converts a gRPC CompVersionRequest into a ComponentVersionsInput DTO.
// It marshals the gRPC request to JSON and then unmarshals it into the internal DTO format.
//
// Parameters:
//   - s: Sugared logger for error logging
//   - request: gRPC component version request
//
// Returns:
//   - ComponentVersionsInput DTO or BadRequestError if conversion fails
func convertCompVersionsInput(s *zap.SugaredLogger, request *pb.CompVersionRequest) (dtos.ComponentVersionsInput, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return dtos.ComponentVersionsInput{}, se.NewBadRequestError("Error parsing request data", err)
	}
	dtoRequest, err := dtos.ParseComponentVersionsInput(s, data)
	if err != nil {
		return dtos.ComponentVersionsInput{}, se.NewBadRequestError("Error parsing request data", err)
	}
	return dtoRequest, nil
}

// convertCompVersionsOutput converts a ComponentVersionsOutput DTO into a gRPC CompVersionResponse.
// It marshals the DTO to JSON and then unmarshals it into the gRPC response format.
//
// Parameters:
//   - s: Sugared logger for error logging
//   - output: ComponentVersionsOutput DTO containing component version data
//
// Returns:
//   - gRPC CompVersionResponse or error if conversion fails
func convertCompVersionsOutput(s *zap.SugaredLogger, output dtos.ComponentVersionsOutput) (*pb.CompVersionResponse, error) {
	data, err := json.Marshal(output)
	if err != nil {
		s.Errorf("Problem marshalling component request output: %v", err)
		return &pb.CompVersionResponse{}, errors.New("problem marshalling component version output")
	}
	var compResp pb.CompVersionResponse
	err = json.Unmarshal(data, &compResp)
	if err != nil {
		s.Errorf("Problem unmarshalling component request output: %v", err)
		return &pb.CompVersionResponse{}, errors.New("problem unmarshalling component version output")
	}
	return &compResp, nil
}

// convertComponentStatusInput converts a gRPC component status request into a ComponentStatusInput DTO.
// It accepts an interface{} to support both REST and gRPC request formats.
// It marshals the request to JSON and then unmarshals it into the internal DTO format.
//
// Parameters:
//   - s: Sugared logger for error logging
//   - request: Generic request interface (gRPC or REST format)
//
// Returns:
//   - ComponentStatusInput DTO or BadRequestError if conversion fails
func convertComponentStatusInput(s *zap.SugaredLogger, request interface{}) (dtos.ComponentStatusInput, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return dtos.ComponentStatusInput{}, se.NewBadRequestError("Error parsing request data", err)
	}
	dtoRequest, err := dtos.ParseComponentStatusInput(s, data)
	if err != nil {
		return dtos.ComponentStatusInput{}, se.NewBadRequestError("Error parsing request data", err)
	}
	return dtoRequest, nil
}

// convertComponentStatusOutput converts a ComponentStatusOutput DTO into a gRPC ComponentStatusResponse.
// It manually constructs the gRPC response structure, handling both success and error cases.
// The function handles optional fields like StatusChangeDate and VersionStatus appropriately.
//
// Parameters:
//   - s: Sugared logger for error logging
//   - output: ComponentStatusOutput DTO containing component and version status data
//
// Returns:
//   - gRPC ComponentStatusResponse with populated fields, or error if conversion fails
func convertComponentStatusOutput(output dtos.ComponentStatusOutput) *pb.ComponentStatusResponse {
	response := &pb.ComponentStatusResponse{
		Purl:        output.Purl,
		Requirement: output.Requirement,
	}
	if output.ComponentStatus.ErrorCode == nil {
		response.ComponentStatus = &pb.ComponentStatusResponse_ComponentStatus{
			FirstIndexedDate: output.ComponentStatus.FirstIndexedDate,
			LastIndexedDate:  output.ComponentStatus.LastIndexedDate,
			Status:           output.ComponentStatus.Status,
			RepositoryStatus: output.ComponentStatus.RepositoryStatus,
		}
		if output.ComponentStatus.StatusChangeDate != "" {
			response.ComponentStatus.StatusChangeDate = output.ComponentStatus.StatusChangeDate
		}
		response.Name = output.Name
	} else {
		response.ComponentStatus = &pb.ComponentStatusResponse_ComponentStatus{
			ErrorMessage: output.ComponentStatus.ErrorMessage,
			ErrorCode:    domain.StatusCodeToErrorCode(*output.ComponentStatus.ErrorCode),
		}
		return response
	}
	if output.VersionStatus != nil {
		if output.VersionStatus.ErrorCode == nil {
			response.VersionStatus = &pb.ComponentStatusResponse_VersionStatus{
				Version:          output.VersionStatus.Version,
				RepositoryStatus: output.VersionStatus.RepositoryStatus,
				Status:           output.VersionStatus.Status,
				IndexedDate:      output.VersionStatus.IndexedDate,
			}
			if output.VersionStatus.StatusChangeDate != "" {
				response.VersionStatus.StatusChangeDate = output.VersionStatus.StatusChangeDate
			}
		} else {
			response.VersionStatus = &pb.ComponentStatusResponse_VersionStatus{
				Version:      output.VersionStatus.Version,
				ErrorMessage: output.VersionStatus.ErrorMessage,
				ErrorCode:    domain.StatusCodeToErrorCode(*output.VersionStatus.ErrorCode),
			}
		}
	}

	return response
}

// convertComponentsStatusInput converts a gRPC components status request into a ComponentsStatusInput DTO.
// It accepts an interface{} to support both REST and gRPC request formats for batch status requests.
// It marshals the request to JSON and then unmarshals it into the internal DTO format.
//
// Parameters:
//   - s: Sugared logger for error logging
//   - request: Generic request interface (gRPC or REST format) containing multiple component status requests
//
// Returns:
//   - ComponentsStatusInput DTO or BadRequestError if conversion fails
func convertComponentsStatusInput(s *zap.SugaredLogger, request interface{}) (dtos.ComponentsStatusInput, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return dtos.ComponentsStatusInput{}, se.NewBadRequestError("Error parsing request data", err)
	}
	dtoRequest, err := dtos.ParseComponentsStatusInput(s, data)
	if err != nil {
		return dtos.ComponentsStatusInput{}, se.NewBadRequestError("Error parsing request data", err)
	}
	return dtoRequest, nil
}

// convertComponentsStatusOutput converts a ComponentsStatusOutput DTO into a gRPC ComponentsStatusResponse.
// It iterates through multiple component status results and converts each one using convertComponentStatusOutput.
// This function handles batch status responses for multiple components.
//
// Parameters:
//   - s: Sugared logger for error logging
//   - output: ComponentsStatusOutput DTO containing multiple component status results
//
// Returns:
//   - gRPC ComponentsStatusResponse with all converted component status entries, or error if conversion fails
func convertComponentsStatusOutput(output dtos.ComponentsStatusOutput) *pb.ComponentsStatusResponse {
	var statusResp pb.ComponentsStatusResponse
	for _, c := range output.Components {
		cs := convertComponentStatusOutput(c)
		statusResp.Components = append(statusResp.Components, cs)
	}
	return &statusResp
}
