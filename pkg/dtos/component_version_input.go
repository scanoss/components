package dtos

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
)

type ComponentVersionsInput struct {
	Purl  string `json:"purl"`
	Limit int    `json:"limit"`
}

func ExportComponentVersionsInput(s *zap.SugaredLogger, output ComponentVersionsInput) ([]byte, error) {
	data, err := json.Marshal(output)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON ")
	}
	return data, nil
}

func ParseComponentVersionsInput(s *zap.SugaredLogger, input []byte) (ComponentVersionsInput, error) {
	if len(input) == 0 {
		return ComponentVersionsInput{}, errors.New("no data supplied to parse")
	}
	var data ComponentVersionsInput
	err := json.Unmarshal(input, &data)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return ComponentVersionsInput{}, fmt.Errorf("failed to parse data: %v", err)
	}
	return data, nil
}
