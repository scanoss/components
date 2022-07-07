package dtos

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	zlog "scanoss.com/components/pkg/logger"
	"testing"
)

func TestParseComponentVersionsInput(t *testing.T) {

	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()

	goodTest := []struct {
		input string
		want  ComponentVersionsInput
	}{
		{
			input: `{"purl": "pkg:npm/scanoss/scanoss.js", "limit": 30}`,
			want:  ComponentVersionsInput{Purl: "pkg:npm/scanoss/scanoss.js", Limit: 30},
		},
		{
			input: `{"purl": "pkg:npm/angular"}`,
			want:  ComponentVersionsInput{Purl: "pkg:npm/angular"},
		},
		{
			input: `{}`,
			want:  ComponentVersionsInput{},
		},
	}

	badTest := []struct {
		input       string
		want        ComponentVersionsInput
		description string
	}{
		{
			description: "Broken JSON, not includes a comma ",
			input:       `{"purl": "pkg:npm/scanoss/scanoss.js" "limit": 30}`,
		},
		{
			description: "purl with number instead of string",
			input:       `{"purl": 99}`,
		},
	}

	for _, test := range goodTest {
		res, err := ParseComponentVersionsInput([]byte(test.input))
		if (!cmp.Equal(test.want, res)) || (err != nil) {
			t.Errorf("Error testing dto: %v\n. Wanted %v, Input: %v \n", err, test.want, test.input)
		}
	}

	// All the test in this table are expected to fail
	for _, test := range badTest {
		if _, err := ParseComponentVersionsInput([]byte(test.input)); err == nil {
			t.Errorf("Expected an error for input: %v", test.input)
		}
	}

	_, err = ParseComponentVersionsInput([]byte(""))
	if err == nil {
		t.Errorf("Expected an error for empty input")
	}

	_, err = ParseComponentVersionsInput(nil)
	if err == nil {
		t.Errorf("Expected an error for empty input")
	}
}

func TestExportComponentVersionsInput(t *testing.T) {

	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()

	bytes, err := ExportComponentVersionsInput(ComponentVersionsInput{Purl: "pkg:npm/scanoss/scanoss.js"})
	if err != nil {
		t.Errorf("Failed to export component version input: %v\n", err)
	}
	fmt.Printf("Converting component version input json to bytes: %v\n", bytes)

	bytes, err = ExportComponentVersionsInput(ComponentVersionsInput{})
	if err != nil {
		t.Errorf("Failed to export component version input: %v\n", err)
	}
	fmt.Printf("Converting empty component version input json to bytes: %v\n", bytes)

}
