package service

import (
	"encoding/json"
	"errors"
	pb "github.com/scanoss/papi/api/componentsv2"
	"scanoss.com/components/pkg/dtos"
	zlog "scanoss.com/components/pkg/logger"
)

func convertSearchComponentInput(request *pb.CompSearchRequest) (dtos.ComponentSearchInput, error) {
	data, err := json.Marshal(request)
	if err != nil {
		zlog.S.Errorf("Problem marshalling component request input: %v", err)
		return dtos.ComponentSearchInput{}, errors.New("problem marshalling component input")
	}
	dtoRequest, err := dtos.ParseComponentSearchInput(data)
	if err != nil {
		zlog.S.Errorf("Problem parsing component request input: %v", err)
		return dtos.ComponentSearchInput{}, errors.New("problem parsing component input")
	}
	return dtoRequest, nil
}

func convertSearchComponentOutput(output dtos.ComponentsSearchOutput) (*pb.CompSearchResponse, error) {
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

func convertCompVersionsInput(request *pb.CompVersionRequest) (dtos.ComponentVersionsInput, error) {
	data, err := json.Marshal(request)
	if err != nil {
		zlog.S.Errorf("Problem marshalling component request input: %v", err)
		return dtos.ComponentVersionsInput{}, errors.New("problem marshalling component version request input")
	}
	dtoRequest, err := dtos.ParseComponentVersionsInput(data)
	if err != nil {
		zlog.S.Errorf("Problem parsing component request input: %v", err)
		return dtos.ComponentVersionsInput{}, errors.New("problem parsing component version input")
	}
	return dtoRequest, nil
}

func convertCompVersionsOutput(output dtos.ComponentVersionsOutput) (*pb.CompVersionResponse, error) {
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
