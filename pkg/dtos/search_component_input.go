package dtos

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
)

type ComponentSearchInput struct {
	Search    string `json:"search"`
	Vendor    string `json:"vendor" `
	Component string `json:"component"`
	Package   string `json:"package"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

func ParseComponentSearchInput(s *zap.SugaredLogger, input []byte) (ComponentSearchInput, error) {
	if len(input) == 0 {
		return ComponentSearchInput{}, errors.New("no data supplied to parse")
	}
	var data ComponentSearchInput
	err := json.Unmarshal(input, &data)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return ComponentSearchInput{}, fmt.Errorf("failed to parse data: %v", err)
	}
	return data, nil
}

func ExportComponentSearchInput(s *zap.SugaredLogger, input ComponentSearchInput) ([]byte, error) {
	data, err := json.Marshal(input)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON ")
	}
	return data, nil
}
