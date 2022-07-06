package dtos

import (
	"encoding/json"
	"errors"
	"fmt"
	zlog "scanoss.com/components/pkg/logger"
)

type ComponentsSearchOutput struct {
	Components []ComponentSearchOutput `json:"components"`
}

type ComponentSearchOutput struct {
	Component string `json:"component"`
	Purl      string `json:"purl"`
	Url       string `json:"url"`
}

func ExportComponentSearchOutput(output ComponentsSearchOutput) ([]byte, error) {
	data, err := json.Marshal(output)
	if err != nil {
		zlog.S.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to produce JSON ")
	}
	return data, nil
}

func ParseComponentSearchOutput(input []byte) (ComponentsSearchOutput, error) {
	if input == nil || len(input) == 0 {
		return ComponentsSearchOutput{}, errors.New("no data supplied to parse")
	}
	var data ComponentsSearchOutput
	err := json.Unmarshal(input, &data)
	if err != nil {
		zlog.S.Errorf("Parse failure: %v", err)
		return ComponentsSearchOutput{}, errors.New(fmt.Sprintf("failed to parse data: %v", err))
	}
	zlog.S.Debugf("Parsed data: %v", data)
	return data, nil
}
