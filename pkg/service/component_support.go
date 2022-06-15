package service

import (
	"encoding/json"
	"errors"
	pb "github.com/scanoss/papi/api/componentsv2"
	"scanoss.com/components/pkg/dtos/dtoGetComponentVersion"
	"scanoss.com/components/pkg/dtos/dtoSearchComponent"
	zlog "scanoss.com/components/pkg/logger"
)

func convertSearchComponentInput(request *pb.CompSearchRequest) (dtoSearchComponent.ComponentSearchInput, error) {
	data, err := json.Marshal(request)
	if err != nil {
		zlog.S.Errorf("Problem marshalling component request input: %v", err)
		return dtoSearchComponent.ComponentSearchInput{}, errors.New("problem marshalling component input")
	}
	dtoRequest, err := dtoSearchComponent.ParseComponentInput(data)
	if err != nil {
		zlog.S.Errorf("Problem parsing component request input: %v", err)
		return dtoSearchComponent.ComponentSearchInput{}, errors.New("problem parsing component input")
	}
	return dtoRequest, nil
}

func convertSearchComponentOutput(output dtoSearchComponent.ComponentsSearchOutput) (*pb.CompSearchResponse, error) {
	data, err := json.Marshal(output)
	if err != nil {
		zlog.S.Errorf("Problem marshalling component request output: %v", err)
		return &pb.CompSearchResponse{}, errors.New("problem marshalling component output")
	}
	zlog.S.Debugf("Parsed data: %v", string(data))
	var compResp pb.CompSearchResponse
	err = json.Unmarshal(data, &compResp)
	if err != nil {
		zlog.S.Errorf("Problem unmarshalling component request output: %v", err)
		return &pb.CompSearchResponse{}, errors.New("problem unmarshalling component output")
	}
	return &compResp, nil
}

func convertCompVersionsInput(request *pb.CompVersionRequest) (dtoGetComponentVersion.ComponentVersionsInput, error) {
	data, err := json.Marshal(request)
	if err != nil {
		zlog.S.Errorf("Problem marshalling component request input: %v", err)
		return dtoGetComponentVersion.ComponentVersionsInput{}, errors.New("problem marshalling component version request input")
	}
	dtoRequest, err := dtoGetComponentVersion.ParseComponentVersionsInput(data)
	if err != nil {
		zlog.S.Errorf("Problem parsing component request input: %v", err)
		return dtoGetComponentVersion.ComponentVersionsInput{}, errors.New("problem parsing component version input")
	}
	return dtoRequest, nil
}

func convertCompVersionsOutput(output dtoGetComponentVersion.ComponentVersionsOutput) (*pb.CompVersionResponse, error) {
	data, err := json.Marshal(output)
	if err != nil {
		zlog.S.Errorf("Problem marshalling component request output: %v", err)
		return &pb.CompVersionResponse{}, errors.New("problem marshalling component version output")
	}
	zlog.S.Debugf("Parsed data: %v", string(data))
	var compResp pb.CompVersionResponse
	err = json.Unmarshal(data, &compResp)
	if err != nil {
		zlog.S.Errorf("Problem unmarshalling component request output: %v", err)
		return &pb.CompVersionResponse{}, errors.New("problem unmarshalling component version output")
	}
	return &compResp, nil
}
