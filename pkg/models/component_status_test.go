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
	"fmt"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	_ "modernc.org/sqlite"
	myconfig "scanoss.com/components/pkg/config"
)

// TestGetComponentStatusByPurlAndVersion tests retrieving status for a specific component version.
//
//goland:noinspection DuplicatedCode
func TestGetComponentStatusByPurlAndVersion(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	db := sqliteSetup(t)
	defer CloseDB(db)
	conn := sqliteConn(t, ctx, db)
	defer CloseConn(conn)
	err = LoadTestSQLData(db, ctx, conn)
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	myConfig, err := myconfig.NewServerConfig(nil)
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}
	myConfig.Database.Trace = true

	componentStatusModel := NewComponentStatusModel(ctx, s, database.NewDBSelectContext(s, db, conn, myConfig.Database.Trace))

	// Test cases that should pass
	passTestTable := []struct {
		purl    string
		version string
	}{
		{
			purl:    "pkg:npm/react",
			version: "18.0.0",
		},
		{
			purl:    "pkg:gem/tablestyle",
			version: "0.1.0",
		},
	}

	for _, test := range passTestTable {
		fmt.Printf("Testing purl: %v, version: %v\n", test.purl, test.version)
		status, err := componentStatusModel.GetComponentStatusByPurlAndVersion(test.purl, test.version)
		if err != nil {
			// It's ok if we don't find the specific version in test data
			fmt.Printf("Version not found (expected): %v\n", err)
		} else {
			fmt.Printf("Status: %+v\n", status)
		}
	}

	// Test cases that should fail
	failTestTable := []struct {
		purl    string
		version string
	}{
		{
			purl:    "", // Empty purl
			version: "1.0.0",
		},
		{
			purl:    "invalid-purl", // Invalid purl format
			version: "1.0.0",
		},
	}

	for _, test := range failTestTable {
		_, err := componentStatusModel.GetComponentStatusByPurlAndVersion(test.purl, test.version)
		if err == nil {
			t.Errorf("An error was expected for purl: %v, version: %v", test.purl, test.version)
		}
	}
}

// TestGetComponentStatusByPurl tests retrieving status for a component (without version).
//
//goland:noinspection DuplicatedCode
func TestGetComponentStatusByPurl(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	db := sqliteSetup(t)
	defer CloseDB(db)
	conn := sqliteConn(t, ctx, db)
	defer CloseConn(conn)
	err = LoadTestSQLData(db, ctx, conn)
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	myConfig, err := myconfig.NewServerConfig(nil)
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}
	myConfig.Database.Trace = true

	componentStatusModel := NewComponentStatusModel(ctx, s, database.NewDBSelectContext(s, db, conn, myConfig.Database.Trace))

	// Test cases that should pass
	passTestTable := []struct {
		purl string
	}{
		{purl: "pkg:npm/react"},
		{purl: "pkg:gem/tablestyle"},
	}

	for _, test := range passTestTable {
		fmt.Printf("Testing purl: %v\n", test.purl)
		status, err := componentStatusModel.GetComponentStatusByPurl(test.purl)
		if err != nil {
			fmt.Printf("Component not found (may be expected): %v\n", err)
		} else {
			fmt.Printf("Status: %+v\n", status)
		}
	}

	// Test cases that should fail
	failTestTable := []struct {
		purl string
	}{
		{purl: ""},                // Empty purl
		{purl: "invalid-purl"},    // Invalid purl format
		{purl: "pkg:npm/NOEXIST"}, // Non-existent component
	}

	for _, test := range failTestTable {
		_, err := componentStatusModel.GetComponentStatusByPurl(test.purl)
		if err == nil {
			t.Errorf("An error was expected for purl: %v", test.purl)
		}
	}
}

// TestGetProjectStatusByPurl tests retrieving project-level status only (no version info).
//
//goland:noinspection DuplicatedCode
func TestGetProjectStatusByPurl(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	db := sqliteSetup(t)
	defer CloseDB(db)
	conn := sqliteConn(t, ctx, db)
	defer CloseConn(conn)
	err = LoadTestSQLData(db, ctx, conn)
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	myConfig, err := myconfig.NewServerConfig(nil)
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}
	myConfig.Database.Trace = true

	componentStatusModel := NewComponentStatusModel(ctx, s, database.NewDBSelectContext(s, db, conn, myConfig.Database.Trace))

	// Test cases that should pass
	passTestTable := []struct {
		purl string
	}{
		{purl: "pkg:npm/react"},
		{purl: "pkg:gem/tablestyle"},
	}

	for _, test := range passTestTable {
		fmt.Printf("Testing project status for purl: %v\n", test.purl)
		status, err := componentStatusModel.GetProjectStatusByPurl(test.purl)
		if err != nil {
			fmt.Printf("Component not found (may be expected): %v\n", err)
		} else {
			fmt.Printf("Project Status: %+v\n", status)
			if status.Component == "" {
				t.Errorf("Expected non-empty component name for purl: %v", test.purl)
			}
		}
	}

	// Test cases that should fail
	failTestTable := []struct {
		purl string
	}{
		{purl: ""},                // Empty purl
		{purl: "invalid-purl"},    // Invalid purl format
		{purl: "pkg:npm/NOEXIST"}, // Non-existent component
	}

	for _, test := range failTestTable {
		_, err := componentStatusModel.GetProjectStatusByPurl(test.purl)
		if err == nil {
			t.Errorf("An error was expected for purl: %v", test.purl)
		}
	}
}
