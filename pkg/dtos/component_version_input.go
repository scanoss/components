package dtos

import (
	"encoding/json"
	"errors"
	"fmt"
	zlog "scanoss.com/components/pkg/logger"
)

type ComponentVersionsInput struct {
	Purl  string `json:"purl"`
	Limit int    `json:"limit"`
}

func ExportComponentVersionsInput(output ComponentVersionsInput) ([]byte, error) {
	data, err := json.Marshal(output)
	if err != nil {
		zlog.S.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON ")
	}
	return data, nil
}

func ParseComponentVersionsInput(input []byte) (ComponentVersionsInput, error) {
	if input == nil || len(input) == 0 {
		return ComponentVersionsInput{}, errors.New("no data supplied to parse")
	}
	var data ComponentVersionsInput
	err := json.Unmarshal(input, &data)
	if err != nil {
		zlog.S.Errorf("Parse failure: %v", err)
		return ComponentVersionsInput{}, errors.New(fmt.Sprintf("failed to parse data: %v", err))
	}
	zlog.S.Debugf("Parsed data: %v", data)
	return data, nil
}
