package dtos

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	zlog "scanoss.com/components/pkg/logger"
	"testing"
)

func TestParseComponentSearchOutput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()

	goodTest := []struct {
		input string
		want  ComponentsSearchOutput
	}{
		{
			input: `{"components": [{"component": "@angular/elements", "purl": "pkg:npm/%40angular/elements", "url": "https://www.npmjs.com/package/%40angular/elements"}]}`,
			want: ComponentsSearchOutput{Components: []ComponentSearchOutput{
				{
					Component: "@angular/elements",
					Purl:      "pkg:npm/%40angular/elements",
					Url:       "https://www.npmjs.com/package/%40angular/elements",
				},
			}},
		},
		{
			input: `{"components": [{"component": "angular", "purl": "pkg:npm/angular" }]}`,
			want: ComponentsSearchOutput{Components: []ComponentSearchOutput{
				{
					Component: "angular",
					Purl:      "pkg:npm/angular",
				},
			}},
		},
	}

	badTest := []struct {
		input       string
		description string
	}{
		{
			description: "Broken JSON, bad type ",
			input:       `{"components": ["a"]}`,
		},
	}

	for _, test := range goodTest {
		res, err := ParseComponentSearchOutput([]byte(test.input))
		if !cmp.Equal(test.want, res) || err != nil {
			t.Errorf("Error generating dto: %v\n. Wanted %v\n, Got: %v \n", err, test.want, res)
		}
	}

	// All the test in this table are expected to fail
	for _, test := range badTest {
		if _, err := ParseComponentSearchOutput([]byte(test.input)); err == nil {
			t.Errorf("Expected an error for input: %v", test.input)
		}
	}

	_, err = ParseComponentSearchOutput([]byte(""))
	if err == nil {
		t.Errorf("Expected an error for empty input")
	}
}

func TestExportComponentSearchOutput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()

	fullComponent := ComponentsSearchOutput{Components: []ComponentSearchOutput{
		{
			Component: "@angular/elements",
			Purl:      "pkg:npm/%40angular/elements",
			Url:       "https://www.npmjs.com/package/%40angular/elements",
		},
	}}

	data, err := ExportComponentSearchOutput(fullComponent)
	if err != nil {
		t.Errorf("ExportComponentSearchOutput() error = %v", err)
	}
	fmt.Println("Exported output data: ", data)

	data, err = ExportComponentSearchOutput(ComponentsSearchOutput{})
	if err != nil {
		t.Errorf("ExportComponentSearchOutput() error = %v", err)
	}
	fmt.Println("Exported output data: ", data)

}
