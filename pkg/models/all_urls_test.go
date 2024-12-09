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

package models

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	myconfig "scanoss.com/components/pkg/config"
	"testing"
)

// setupTest initializes all necessary components for testing
func setupTest(t *testing.T) (*sqlx.DB, *sqlx.Conn, *AllUrlsModel) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	db := sqliteSetup(t)           // Setup SQL Lite DB
	conn := sqliteConn(t, ctx, db) // Get a connection from the pool
	err = LoadTestSQLData(db, ctx, conn)
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	myConfig, err := myconfig.NewServerConfig(nil)
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}
	myConfig.Database.Trace = true

	return db, conn, NewAllUrlModel(ctx, s, database.NewDBSelectContext(s, db, conn, myConfig.Database.Trace))
}

// cleanup handles proper resource cleanup
func cleanup(db *sqlx.DB, conn *sqlx.Conn) {
	CloseConn(conn)
	CloseDB(db)
	zlog.SyncZap()
}

// TestGetUrlsByPurlNameType tests the GetUrlsByPurlNameType function
func TestGetUrlsByPurlNameType(t *testing.T) {
	db, conn, allUrlsModel := setupTest(t)
	defer cleanup(db, conn)

	tests := []struct {
		name      string
		purlName  string
		purlType  string
		limit     int
		wantErr   bool
		wantEmpty bool
		validate  func(t *testing.T, urls []AllUrl)
	}{
		{
			name:      "valid url search",
			purlName:  "tablestyle",
			purlType:  "gem",
			limit:     -1,
			wantErr:   false,
			wantEmpty: false,
			validate: func(t *testing.T, urls []AllUrl) {
				if urls[0].PurlName != "tablestyle" {
					t.Errorf("expected purlName 'tablestyle', got %s", urls[0].PurlName)
				}
			},
		},
		{
			name:      "grpcio_1_12_1 has MIT and Apache licenses",
			purlName:  "grpcio",
			purlType:  "pypi",
			limit:     200,
			wantErr:   false,
			wantEmpty: false,
			validate: func(t *testing.T, urls []AllUrl) {
				// Filter URLs for specific version
				var matchedUrls []AllUrl
				for _, url := range urls {
					if url.PurlName == "grpcio" && url.Version == "1.12.1" {
						matchedUrls = append(matchedUrls, url)
					}
				}

				if len(matchedUrls) == 0 {
					t.Errorf("no URLs found for grpcio version 1.12.1")
					return
				}

				// Create a map to track found licenses
				foundLicenses := make(map[string]bool)
				for _, url := range matchedUrls {
					foundLicenses[url.License] = true
				}

				// Check for required licenses
				requiredLicenses := []string{"MIT", "Apache License 2.0"}
				for _, license := range requiredLicenses {
					if !foundLicenses[license] {
						t.Errorf("expected license %s not found for grpcio 1.12.1", license)
					}
				}
			},
		},
		{
			name:      "nonexistent url",
			purlName:  "NONEXISTENT",
			purlType:  "none",
			limit:     -1,
			wantErr:   false,
			wantEmpty: true,
			validate:  nil,
		},
		{
			name:      "empty purlName",
			purlName:  "",
			purlType:  "gem",
			limit:     -1,
			wantErr:   true,
			wantEmpty: true,
			validate:  nil,
		},
		{
			name:      "empty purlType",
			purlName:  "tablestyle",
			purlType:  "",
			limit:     -1,
			wantErr:   true,
			wantEmpty: true,
			validate:  nil,
		},
		{
			name:      "zero limit",
			purlName:  "tablestyle",
			purlType:  "gem",
			limit:     0,
			wantErr:   false,
			wantEmpty: false,
			validate:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls, err := allUrlsModel.GetUrlsByPurlNameType(tt.purlName, tt.purlType, tt.limit)

			// Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUrlsByPurlNameType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check empty condition
			if tt.wantEmpty && len(urls) > 0 {
				t.Errorf("GetUrlsByPurlNameType() expected empty results, got %v", urls)
			}

			// Run custom validations if provided
			if tt.validate != nil {
				tt.validate(t, urls)
			}
		})
	}
}

// TestGetUrlsByPurlString tests the GetUrlsByPurlString function
func TestGetUrlsByPurlString(t *testing.T) {
	db, conn, allUrlsModel := setupTest(t)
	defer cleanup(db, conn)

	tests := []struct {
		name       string
		purlString string
		limit      int
		wantErr    bool
		wantEmpty  bool
		validate   func(t *testing.T, urls []AllUrl)
	}{
		{
			name:       "valid purl",
			purlString: "pkg:gem/tablestyle",
			limit:      -1,
			wantErr:    false,
			wantEmpty:  false,
			validate: func(t *testing.T, urls []AllUrl) {
				if urls[0].PurlName != "tablestyle" {
					t.Errorf("expected purlName 'tablestyle', got %s", urls[0].PurlName)
				}
			},
		},
		{
			name:       "empty purl",
			purlString: "",
			limit:      -1,
			wantErr:    true,
			wantEmpty:  true,
			validate:   nil,
		},
		{
			name:       "invalid purl format",
			purlString: "pkg::pypi",
			limit:      -1,
			wantErr:    true,
			wantEmpty:  true,
			validate:   nil,
		},
		{
			name:       "zero limit",
			purlString: "pkg:gem/tablestyle",
			limit:      0,
			wantErr:    false,
			wantEmpty:  false,
			validate:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls, err := allUrlsModel.GetUrlsByPurlString(tt.purlString, tt.limit)

			// Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUrlsByPurlString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check empty condition
			if tt.wantEmpty && len(urls) > 0 {
				t.Errorf("GetUrlsByPurlString() expected empty results, got %v", urls)
			}

			// Run custom validations if provided
			if tt.validate != nil {
				tt.validate(t, urls)
			}
		})
	}
}
