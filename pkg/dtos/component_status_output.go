package dtos

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"go.uber.org/zap"
)

// ComponentStatusOutput represents the status information for a single component
type ComponentStatusOutput struct {
	Purl            string               `json:"purl"`
	Name            string               `json:"name"`
	Requirement     string               `json:"requirement,omitempty"`
	VersionStatus   *VersionStatusOutput `json:"version_status,omitempty"`
	ComponentStatus *ComponentStatusInfo `json:"component_status,omitempty"`
}

// VersionStatusOutput represents the status of a specific version
type VersionStatusOutput struct {
	Version          string             `json:"version"`
	Status           string             `json:"status"`
	RepositoryStatus string             `json:"repository_status,omitempty"`
	IndexedDate      string             `json:"indexed_date,omitempty"`
	StatusChangeDate string             `json:"status_change_date,omitempty"`
	ErrorMessage     *string            `json:"error_message,omitempty"`
	ErrorCode        *domain.StatusCode `json:"error_code,omitempty"`
}

// ComponentStatusInfo represents the status of a component (ignoring version)
type ComponentStatusInfo struct {
	Status           string             `json:"status"`
	RepositoryStatus string             `json:"repository_status,omitempty"`
	FirstIndexedDate string             `json:"first_indexed_date,omitempty"`
	LastIndexedDate  string             `json:"last_indexed_date,omitempty"`
	StatusChangeDate string             `json:"status_change_date,omitempty"`
	ErrorMessage     *string            `json:"error_message,omitempty"`
	ErrorCode        *domain.StatusCode `json:"error_code,omitempty"`
}

// ComponentsStatusOutput represents the status information for multiple components
type ComponentsStatusOutput struct {
	Components []ComponentStatusOutput `json:"components"`
}

func ExportComponentStatusOutput(s *zap.SugaredLogger, output ComponentStatusOutput) ([]byte, error) {
	data, err := json.Marshal(output)
	if err != nil {
		s.Errorf("Marshal failure: %v", err)
		return nil, errors.New("failed to produce JSON")
	}
	return data, nil
}

func ParseComponentStatusOutput(s *zap.SugaredLogger, output []byte) (ComponentStatusOutput, error) {
	if len(output) == 0 {
		return ComponentStatusOutput{}, errors.New("no data supplied to parse")
	}
	var data ComponentStatusOutput
	err := json.Unmarshal(output, &data)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return ComponentStatusOutput{}, fmt.Errorf("failed to parse data: %v", err)
	}
	return data, nil
}

func ExportComponentsStatusOutput(s *zap.SugaredLogger, output ComponentsStatusOutput) ([]byte, error) {
	data, err := json.Marshal(output)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON ")
	}
	return data, nil
}

func ParseComponentsStatusOutput(s *zap.SugaredLogger, output []byte) (ComponentsStatusOutput, error) {
	if len(output) == 0 {
		return ComponentsStatusOutput{}, errors.New("no data supplied to parse")
	}
	var data ComponentsStatusOutput
	err := json.Unmarshal(output, &data)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return ComponentsStatusOutput{}, fmt.Errorf("failed to parse data: %v", err)
	}
	return data, nil
}

// StringPtr returns a pointer to the provided string value
// This is useful for optional string fields that require pointers
func StringPtr(s string) *string {
	return &s
}
