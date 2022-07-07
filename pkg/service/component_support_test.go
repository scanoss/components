package service

import (
	"fmt"
	pb "github.com/scanoss/papi/api/componentsv2"
	"scanoss.com/components/pkg/dtos"
	zlog "scanoss.com/components/pkg/logger"
	"testing"
)

func TestConvertSearchComponentInput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()

	compSearchRequest := new(pb.CompSearchRequest)
	compSearchRequest.Search = "angular"
	compSearchRequest.Package = "github"

	dto, err := convertSearchComponentInput(compSearchRequest)
	if err != nil {
		t.Errorf("Error generating dto from protobuff request: %v\n", err)
	}
	fmt.Printf("dto component input: %v\n", dto)

}

func TestConvertSearchComponentOutput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()

	dtoOut := dtos.ComponentsSearchOutput{Components: []dtos.ComponentSearchOutput{
		{
			Component: "angular",
			Purl:      "pkg:github/bclinkinbeard/angular",
			Url:       "https://github.com/bclinkinbeard/angular",
		},
	}}

	protobuffSearchOut, err := convertSearchComponentOutput(dtoOut)
	if err != nil {
		t.Errorf("An error ocurred when convertin dto to protobuf %v\n", err)
	}
	fmt.Printf("Protobuff created: %v\n", protobuffSearchOut)
}

func TestConvertCompVersionsInput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()

	compVersionRequest := new(pb.CompVersionRequest)
	compVersionRequest.Purl = "angular"

	dto, err := convertCompVersionsInput(compVersionRequest)
	if err != nil {
		t.Errorf("Error generating dto from protobuff request: %v\n", err)
	}
	fmt.Printf("dto component input: %v\n", dto)
}

func TestConvertCompVersionsOutput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()

	dtoVersionOut := dtos.ComponentVersionsOutput{
		Component: dtos.ComponentOutput{
			Component: "@angular/elements",
			Purl:      "pkg:npm/%40angular/elements",
			Url:       "https://www.npmjs.com/package/%40angular/elements",
			Versions: []dtos.ComponentVersion{
				{
					Version: "1.8.3",
					Licenses: []dtos.ComponentLicense{
						{
							Name:   "MIT",
							SpdxId: "MIT",
							IsSpdx: true,
						},
					},
				},
			},
		},
	}

	protobuffOut, err := convertCompVersionsOutput(dtoVersionOut)
	if err != nil {
		t.Errorf("Error converting dto to protobuff request: %v\n", err)
	}
	fmt.Printf("dto component input: %v\n", protobuffOut)

}
