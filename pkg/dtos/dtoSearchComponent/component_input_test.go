package dtoSearchComponent

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
		want  ComponentSearchInput
	}{
		{
			input: `{"search": "angular", "package": "npm", "limit": 30 }`,
			want:  ComponentSearchInput{Search: "angular", Package: "npm", Limit: 30},
		},
		{
			input: `{"component": "angular", "package": "github", "offset": 0 }`,
			want:  ComponentSearchInput{Component: "angular", Package: "github", Offset: 0},
		},
		{
			input: `{}`,
			want:  ComponentSearchInput{},
		},
		{
			input: `{"vendor": "scanoss" }`,
			want:  ComponentSearchInput{Vendor: "scanoss"},
		},
	}

	badTest := []struct {
		input       string
		want        ComponentSearchInput
		description string
	}{
		{
			description: "Broken JSON, not ending with }",
			input:       `{"search": "angular", "package": "npm", "limit": 30`,
		},
		{
			description: "Vendor with type number instead of string",
			input:       `{"vendor": 23 }`,
		},
	}

	for _, test := range goodTest {
		if res, err := ParseComponentInput([]byte(test.input)); !cmp.Equal(test.want, res) || err != nil {
			if err != nil {
				t.Errorf("Error generating dto: %v\n. Wanted %v, Input: %v \n", err, test.want, test.input)
			}
		}
	}

	// All the test in this table are expected to fail
	for _, test := range badTest {
		if _, err := ParseComponentInput([]byte(test.input)); err == nil {
			t.Errorf("Expected an error for input: %v", test.input)
		}
	}

}
