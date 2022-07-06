package dtos

import (
	"encoding/json"
	"errors"
	"fmt"
	zlog "scanoss.com/components/pkg/logger"
)

type ComponentSearchInput struct {
	Search    string `json:"search"`
	Vendor    string `json:"vendor" `
	Component string `json:"component"`
	Package   string `json:"package"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

func ParseComponentSearchInput(input []byte) (ComponentSearchInput, error) {
	if input == nil || len(input) == 0 {
		return ComponentSearchInput{}, errors.New("no data supplied to parse")
	}
	var data ComponentSearchInput
	err := json.Unmarshal(input, &data)
	if err != nil {
		zlog.S.Errorf("Parse failure: %v", err)
		return ComponentSearchInput{}, errors.New(fmt.Sprintf("failed to parse data: %v", err))
	}
	zlog.S.Debugf("Parsed data2: %v", data)
	return data, nil
}

func ExportComponentSearchInput(input ComponentSearchInput) ([]byte, error) {
	data, err := json.Marshal(input)
	if err != nil {
		zlog.S.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON ")
	}
	return data, nil
}
