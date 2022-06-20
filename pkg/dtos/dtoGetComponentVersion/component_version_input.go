package dtoGetComponentVersion

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

func ParseComponentVersionsInput(input []byte) (ComponentVersionsInput, error) {
	if input == nil || len(input) == 0 {
		return ComponentVersionsInput{}, errors.New("no input component data supplied to parse")
	}
	var data ComponentVersionsInput
	err := json.Unmarshal(input, &data)
	if err != nil {
		zlog.S.Errorf("Parse failure: %v", err)
		return ComponentVersionsInput{}, errors.New(fmt.Sprintf("failed to parse component versions input data: %v", err))
	}
	zlog.S.Debugf("Parsed data: %v", data)
	return data, nil
}
