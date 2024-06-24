package dtos

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"testing"
)

func TestParseComponentVersionsOutput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

	goodTest := []struct {
		input string
		want  ComponentVersionsOutput
	}{
		{
			input: `{ "component": {
						"component": "@angular/elements",
    					"purl": "pkg:npm/%40angular/elements",
    					"url": "https://www.npmjs.com/package/%40angular/elements",
    					"versions": [
				      		{ "version": "1.8.3", "licenses": [{ "name": "MIT", "spdx_id": "MIT", "is_spdx_approved": true }]},
      						{ "version": "1.8.2", "licenses": [{ "name": "MIT", "spdx_id": "MIT", "is_spdx_approved": true }]}
						]}
					}`,
			want: ComponentVersionsOutput{Component: ComponentOutput{
				Component: "@angular/elements",
				Purl:      "pkg:npm/%40angular/elements",
				Url:       "https://www.npmjs.com/package/%40angular/elements",
				Versions: []ComponentVersion{
					{
						Version: "1.8.3",
						Licenses: []ComponentLicense{
							{
								Name:   "MIT",
								SpdxId: "MIT",
								IsSpdx: true,
							},
						},
					},
					{
						Version: "1.8.2",
						Licenses: []ComponentLicense{
							{
								Name:   "MIT",
								SpdxId: "MIT",
								IsSpdx: true,
							},
						},
					},
				},
			},
			},
		},
	}

	badTest := []struct {
		input       string
		description string
	}{
		{
			description: "Broken JSON, bad type ",
			input:       `{"component": "a"}`,
		},
	}

	for _, test := range goodTest {
		res, err := ParseComponentVersionsOutput(s, []byte(test.input))
		if !cmp.Equal(test.want, res) || err != nil {
			t.Errorf("Error generating dto: %v\n. Wanted %v\n, Got: %v \n", err, test.want, res)
		}
	}

	// All the test in this table are expected to fail
	for _, test := range badTest {
		if _, err := ParseComponentVersionsOutput(s, []byte(test.input)); err == nil {
			t.Errorf("Expected an error for input: %v", test.input)
		}
	}

	_, err = ParseComponentVersionsOutput(s, []byte(""))
	if err == nil {
		t.Errorf("Expected an error for empty input")
	}

}

func TestExportComponentVersionsOutput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

	fullComponent := ComponentVersionsOutput{Component: ComponentOutput{
		Component: "@angular/elements",
		Purl:      "pkg:npm/%40angular/elements",
		Url:       "https://www.npmjs.com/package/%40angular/elements",
		Versions: []ComponentVersion{
			{
				Version: "1.8.3",
				Licenses: []ComponentLicense{
					{
						Name:   "MIT",
						SpdxId: "MIT",
						IsSpdx: true,
					},
				},
			},
			{
				Version: "1.8.2",
				Licenses: []ComponentLicense{
					{
						Name:   "MIT",
						SpdxId: "MIT",
						IsSpdx: true,
					},
				},
			},
		},
	}}

	data, err := ExportComponentVersionsOutput(s, fullComponent)
	if err != nil {
		t.Errorf("dtos.ExportComponentVersionsOutput() error = %v", err)
	}
	fmt.Println("Exported output data: ", data)

	data, err = ExportComponentVersionsOutput(s, ComponentVersionsOutput{})
	if err != nil {
		t.Errorf("dtos.ExportComponentVersionsOutput() error = %v", err)
	}
	fmt.Println("Exported output data: ", data)

}
