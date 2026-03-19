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

func convertComponentStatusOutput(s *zap.SugaredLogger, output dtos.ComponentStatusOutput) (*pb.ComponentStatusResponse, error) {

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
		return response, nil
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

	return response, nil
}

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

func convertComponentsStatusOutput(s *zap.SugaredLogger, output dtos.ComponentsStatusOutput) (*pb.ComponentsStatusResponse, error) {

	var statusResp pb.ComponentsStatusResponse
	for _, c := range output.Components {
		cs, _ := convertComponentStatusOutput(s, c)
		statusResp.Components = append(statusResp.Components, cs)
	}

	return &statusResp, nil
}
