package dtos

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
)

// ComponentStatusInput represents a single component status request
type ComponentStatusInput struct {
	Purl        string `json:"purl"`
	Requirement string `json:"requirement,omitempty"`
}

// ComponentsStatusInput represents a request for multiple component statuses
type ComponentsStatusInput struct {
	Components []ComponentStatusInput `json:"components"`
}

func ExportComponentStatusInput(s *zap.SugaredLogger, output ComponentStatusInput) ([]byte, error) {
	data, err := json.Marshal(output)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON ")
	}
	return data, nil
}

func ParseComponentStatusInput(s *zap.SugaredLogger, input []byte) (ComponentStatusInput, error) {
	if len(input) == 0 {
		return ComponentStatusInput{}, errors.New("no data supplied to parse")
	}
	var data ComponentStatusInput
	err := json.Unmarshal(input, &data)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return ComponentStatusInput{}, fmt.Errorf("failed to parse data: %v", err)
	}
	return data, nil
}

func ExportComponentsStatusInput(s *zap.SugaredLogger, output ComponentsStatusInput) ([]byte, error) {
	data, err := json.Marshal(output)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON ")
	}
	return data, nil
}

func ParseComponentsStatusInput(s *zap.SugaredLogger, input []byte) (ComponentsStatusInput, error) {
	if len(input) == 0 {
		return ComponentsStatusInput{}, errors.New("no data supplied to parse")
	}
	var data ComponentsStatusInput
	err := json.Unmarshal(input, &data)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return ComponentsStatusInput{}, fmt.Errorf("failed to parse data: %v", err)
	}
	return data, nil
}
