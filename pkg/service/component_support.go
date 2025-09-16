package service

import (
	"encoding/json"
	"errors"
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
