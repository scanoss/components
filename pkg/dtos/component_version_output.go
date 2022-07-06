package dtos

import (
	"encoding/json"
	"errors"
	"fmt"
	zlog "scanoss.com/components/pkg/logger"
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

func ExportComponentVersionsOutput(output ComponentVersionsOutput) ([]byte, error) {
	data, err := json.Marshal(output)
	if err != nil {
		zlog.S.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON ")
	}
	return data, nil
}

func ParseComponentVersionsOutput(input []byte) (ComponentVersionsOutput, error) {
	if input == nil || len(input) == 0 {
		return ComponentVersionsOutput{}, errors.New("no data supplied to parse")
	}
	var data ComponentVersionsOutput
	err := json.Unmarshal(input, &data)
	if err != nil {
		zlog.S.Errorf("Parse failure: %v", err)
		return ComponentVersionsOutput{}, errors.New(fmt.Sprintf("failed to parse data: %v", err))
	}
	zlog.S.Debugf("Parsed data: %v", data)
	return data, nil
}
