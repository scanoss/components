package dtos

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
)

type ComponentsSearchOutput struct {
	Components []ComponentSearchOutput `json:"components"`
}

type ComponentSearchOutput struct {
	Name      string `json:"name"`
	Component string `json:"component"` // Deprecated. Component and name fields will contain the same data until
	// the component field is removed
	Purl string `json:"purl"`
	Url  string `json:"url"`
}

func ExportComponentSearchOutput(s *zap.SugaredLogger, output ComponentsSearchOutput) ([]byte, error) {
	data, err := json.Marshal(output)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON ")
	}
	return data, nil
}

func ParseComponentSearchOutput(s *zap.SugaredLogger, input []byte) (ComponentsSearchOutput, error) {
	if len(input) == 0 {
		return ComponentsSearchOutput{}, errors.New("no data supplied to parse")
	}
	var data ComponentsSearchOutput
	err := json.Unmarshal(input, &data)
	if err != nil {
		s.Errorf("Parse failure: %v", err)
		return ComponentsSearchOutput{}, fmt.Errorf("failed to parse data: %v", err)
	}
	return data, nil
}
