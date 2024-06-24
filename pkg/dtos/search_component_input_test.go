package dtos

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"testing"
)

func TestParseComponentSearchInput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

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
		res, err := ParseComponentSearchInput(s, []byte(test.input))
		if !cmp.Equal(test.want, res) || err != nil {
			t.Errorf("Error generating dto: %v\n. Wanted %v, Input: %v \n", err, test.want, test.input)
		}
	}

	// All the test in this table are expected to fail
	for _, test := range badTest {
		if _, err := ParseComponentSearchInput(s, []byte(test.input)); err == nil {
			t.Errorf("Expected an error for input: %v", test.input)
		}
	}

	_, err = ParseComponentSearchInput(s, []byte(""))
	if err == nil {
		t.Errorf("Expected an error for empty input")
	}
}

func TestExportComponentSearchInput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

	bytes, err := ExportComponentSearchInput(s, ComponentSearchInput{Search: "angular", Package: "npm", Limit: 10})
	if err != nil {
		t.Errorf("Failed to export component search input: %v\n", err)
	}
	fmt.Printf("Converting component search json to bytes: %v\n", bytes)

	bytes, err = ExportComponentSearchInput(s, ComponentSearchInput{})
	if err != nil {
		t.Errorf("Failed to export component search input: %v\n", err)
	}
	fmt.Printf("Converting empty component search input json to bytes: %v\n", bytes)
}
