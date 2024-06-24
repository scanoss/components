package dtos

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
)

type ComponentVersionsOutput struct {
	Component ComponentOutput `json:"component"`
}

type ComponentOutput struct {
	Component string             `json:"component"`
	Purl      string             `json:"purl"`
	Url       string             `json:"url"`
	Versions  []ComponentVersion `json:"versions"`
}

type ComponentVersion struct {
	Version  string             `json:"version"`
	Licenses []ComponentLicense `json:"licenses"`
}

type ComponentLicense struct {
	Name   string `json:"name"`
	SpdxId string `json:"spdx_id"`
	IsSpdx bool   `json:"is_spdx_approved"`
}

func ExportComponentVersionsOutput(s *zap.SugaredLogger, output ComponentVersionsOutput) ([]byte, error) {
	data, err := json.Marshal(output)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON ")
	}
	return data, nil
}

func ParseComponentVersionsOutput(s *zap.SugaredLogger, output []byte) (ComponentVersionsOutput, error) {
	if len(output) == 0 {
		return ComponentVersionsOutput{}, errors.New("no data supplied to parse")
	}
	var data ComponentVersionsOutput
	err := json.Unmarshal(output, &data)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return ComponentVersionsOutput{}, fmt.Errorf("failed to parse data: %v", err)
	}
	return data, nil
}
