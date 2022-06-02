package service

import (
	"encoding/json"
	"errors"
	pb "github.com/scanoss/papi/api/componentsv2"
	"scanoss.com/dependencies/pkg/dtos"
	zlog "scanoss.com/dependencies/pkg/logger"
)

func convertSearchComponentInput(request *pb.CompSearchRequest) (dtos.ComponentSearchInput, error) {
	data, err := json.Marshal(request)
	if err != nil {
		zlog.S.Errorf("Problem marshalling component request input: %v", err)
		return dtos.ComponentSearchInput{}, errors.New("problem marshalling component input")
	}
	dtoRequest, err := dtos.ParseComponentInput(data)
	if err != nil {
		zlog.S.Errorf("Problem parsing dependency request input: %v", err)
		return dtos.ComponentSearchInput{}, errors.New("problem parsing component input")
	}
	return dtoRequest, nil
}

func convertSearchComponentOutput(output dtos.ComponentsSearchResults) (*pb.CompSearchResponse, error) {
	data, err := json.Marshal(output)
	if err != nil {
		zlog.S.Errorf("Problem marshalling component request output: %v", err)
		return &pb.CompSearchResponse{}, errors.New("problem marshalling component output")
	}
	zlog.S.Debugf("Parsed data: %v", string(data))
	var depResp pb.CompSearchResponse
	err = json.Unmarshal(data, &depResp)
	if err != nil {
		zlog.S.Errorf("Problem unmarshalling component request output: %v", err)
		return &pb.CompSearchResponse{}, errors.New("problem unmarshalling component output")
	}
	return &depResp, nil
}
