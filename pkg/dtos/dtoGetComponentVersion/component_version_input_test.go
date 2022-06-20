package dtoGetComponentVersion

import (
	"github.com/google/go-cmp/cmp"
	zlog "scanoss.com/components/pkg/logger"
	"testing"
)

func TestDependencyInput(t *testing.T) {

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
		if res, err := ParseComponentVersionsInput([]byte(test.input)); !cmp.Equal(test.want, res) || err != nil {
			if err != nil {
				t.Errorf("Error generating dto: %v\n. Wanted %v, Input: %v \n", err, test.want, test.input)
			}
		}
	}

	// All the test in this table are expected to fail
	for _, test := range badTest {
		if _, err := ParseComponentVersionsInput([]byte(test.input)); err == nil {
			t.Errorf("Expected an error for input: %v", test.input)
		}
	}
}
