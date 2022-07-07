// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2022 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package service

import (
	"context"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/componentsv2"
	"reflect"
	zlog "scanoss.com/components/pkg/logger"
	"scanoss.com/components/pkg/models"
	"testing"
)

func TestComponentServer_Echo(t *testing.T) {
	ctx := context.Background()
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseDB(db)
	s := NewComponentServer(db)

	type args struct {
		ctx context.Context
		req *common.EchoRequest
	}
	tests := []struct {
		name    string
		s       pb.ComponentsServer
		args    args
		want    *common.EchoResponse
		wantErr bool
	}{
		{
			name: "Echo",
			s:    s,
			args: args{
				ctx: ctx,
				req: &common.EchoRequest{Message: "Hello there!"},
			},
			want: &common.EchoResponse{Message: "Hello there!"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Echo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.Echo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.Echo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponentServer_SearchComponents(t *testing.T) {
	ctx := context.Background()
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseDB(db)
	db.SetMaxOpenConns(1)
	err = models.LoadTestSqlData(db, nil, nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when loading test data", err)
	}
	s := NewComponentServer(db)

	var compRequestData = `{
  		"component": "angular",
		"package": "github"
	}`

	var compReq = pb.CompSearchRequest{}
	err = json.Unmarshal([]byte(compRequestData), &compReq)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when unmarshalling requestd", err)
	}

	type args struct {
		ctx context.Context
		req *pb.CompSearchRequest
	}
	tests := []struct {
		name    string
		s       pb.ComponentsServer
		args    args
		want    *pb.CompSearchResponse
		wantErr bool
	}{
		{
			name: "Search for angular and purltype github without limit",
			s:    s,
			args: args{
				ctx: ctx,
				req: &compReq,
			},
			want: &pb.CompSearchResponse{Status: &common.StatusResponse{Status: common.StatusCode_SUCCESS, Message: "Success"}},
		},
		{
			name: "Search for a empty request",
			s:    s,
			args: args{
				ctx: ctx,
				req: &pb.CompSearchRequest{},
			},
			want:    &pb.CompSearchResponse{Status: &common.StatusResponse{Status: common.StatusCode_FAILED, Message: "Problems encountered extracting components data"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.SearchComponents(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.SearchComponents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got.Status, tt.want.Status) {
				t.Errorf("service.SearchComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponentServer_GetComponentVersions(t *testing.T) {
	ctx := context.Background()
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseDB(db)
	db.SetMaxOpenConns(1)
	err = models.LoadTestSqlData(db, nil, nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when loading test data", err)
	}
	s := NewComponentServer(db)

	var compVersionRequestData = `{
  		"purl": "pkg:npm/%40angular/elements"
	}`

	var compVersionReq = pb.CompVersionRequest{}
	err = json.Unmarshal([]byte(compVersionRequestData), &compVersionReq)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when unmarshalling requestd", err)
	}

	type args struct {
		ctx context.Context
		req *pb.CompVersionRequest
	}
	tests := []struct {
		name    string
		s       pb.ComponentsServer
		args    args
		want    *pb.CompVersionResponse
		wantErr bool
	}{
		{
			name: "Search for angular and purltype github without limit",
			s:    s,
			args: args{
				ctx: ctx,
				req: &compVersionReq,
			},
			want: &pb.CompVersionResponse{Status: &common.StatusResponse{Status: common.StatusCode_SUCCESS, Message: "Success"}},
		},
		{
			name: "Search for a empty request",
			s:    s,
			args: args{
				ctx: ctx,
				req: &pb.CompVersionRequest{},
			},
			want:    &pb.CompVersionResponse{Status: &common.StatusResponse{Status: common.StatusCode_FAILED, Message: "there is no purl to retrieve component"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetComponentVersions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.GetComponentVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got.Status, tt.want.Status) {
				t.Errorf("service.SearchComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}
