package dtos

import (
	"encoding/json"
	"errors"
	"fmt"
	zlog "scanoss.com/components/pkg/logger"
)

type ComponentSearchInput struct {
	Search    string `json:"search,omitempty"`
	Vendor    string `json:"vendor,omitempty" `
	Component string `json:"component,omitempty"`
	Package   string `json:"package"`
	Limit     string `json:"limit,omitempty"`
	Offset    string `json:"offset,omitempty"`
}

func ParseComponentInput(input []byte) (ComponentSearchInput, error) {
	if input == nil || len(input) == 0 {
		return ComponentSearchInput{}, errors.New("no input component data supplied to parse")
	}
	var data ComponentSearchInput
	err := json.Unmarshal(input, &data)
	if err != nil {
		zlog.S.Errorf("Parse failure: %v", err)
		return ComponentSearchInput{}, errors.New(fmt.Sprintf("failed to parse component input data: %v", err))
	}
	zlog.S.Debugf("Parsed data2: %v", data)
	return data, nil
}
