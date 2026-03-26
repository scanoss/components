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

// ExportComponentStatusInput marshals a ComponentStatusInput struct to JSON bytes.
//
// Parameters:
//   - s: Sugared logger for error logging
//   - output: ComponentStatusInput struct to be marshaled
//
// Returns:
//   - JSON byte array representation of the input, or error if marshaling fails
func ExportComponentStatusInput(s *zap.SugaredLogger, output ComponentStatusInput) ([]byte, error) {
	data, err := json.Marshal(output)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON")
	}
	return data, nil
}

// ParseComponentStatusInput unmarshals JSON bytes into a ComponentStatusInput struct.
//
// Parameters:
//   - s: Sugared logger for error logging
//   - input: JSON byte array to be unmarshaled
//
// Returns:
//   - ComponentStatusInput struct populated from JSON, or error if unmarshaling fails or input is empty
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

// ExportComponentsStatusInput marshals a ComponentsStatusInput struct (containing multiple components) to JSON bytes.
//
// Parameters:
//   - s: Sugared logger for error logging
//   - output: ComponentsStatusInput struct containing an array of component status requests
//
// Returns:
//   - JSON byte array representation of the input, or error if marshaling fails
func ExportComponentsStatusInput(s *zap.SugaredLogger, output ComponentsStatusInput) ([]byte, error) {
	data, err := json.Marshal(output)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON")
	}
	return data, nil
}

// ParseComponentsStatusInput unmarshals JSON bytes into a ComponentsStatusInput struct.
// Used for parsing batch component status requests containing multiple components.
//
// Parameters:
//   - s: Sugared logger for error logging
//   - input: JSON byte array to be unmarshaled
//
// Returns:
//   - ComponentsStatusInput struct with array of component status requests, or error if unmarshaling fails or input is empty
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
