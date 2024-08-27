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

package usecase

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	_ "modernc.org/sqlite"
	myconfig "scanoss.com/components/pkg/config"
	"scanoss.com/components/pkg/dtos"
	"scanoss.com/components/pkg/models"
	"testing"
)

func TestComponentUseCase_SearchComponents(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := context.Background()
	ctx = ctxzap.ToContext(ctx, zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseDB(db)
	err = models.LoadTestSQLData(db, nil, nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when loading test data", err)
	}
	myConfig, err := myconfig.NewServerConfig(nil)
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}

	compUc := NewComponents(ctx, s, db, database.NewDBSelectContext(s, db, nil, myConfig.Database.Trace))

	goodTable := []dtos.ComponentSearchInput{
		{
			Search:  "angular",
			Package: "github",
		},
		{
			Component: "angular",
		},
		{
			Vendor: "angular",
		},
		{
			Component: "angular",
			Vendor:    "angular",
		},
	}

	for _, dtoCompSearchInput := range goodTable {
		searchOut, err := compUc.SearchComponents(dtoCompSearchInput)
		if err != nil {
			t.Fatalf("an error '%s' was not expected when getting components", err)
		}
		fmt.Printf("Search response: %+v\n", searchOut)
	}

}

func TestComponentUseCase_GetComponentVersions(t *testing.T) {

	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := context.Background()
	ctx = ctxzap.ToContext(ctx, zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseDB(db)
	err = models.LoadTestSQLData(db, nil, nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when loading test data", err)
	}
	myConfig, err := myconfig.NewServerConfig(nil)
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}

	compUc := NewComponents(ctx, s, db, database.NewDBSelectContext(s, db, nil, myConfig.Database.Trace))

	goodTable := []dtos.ComponentVersionsInput{
		{
			Purl:  "pkg:gem/tablestyle",
			Limit: 0,
		},
		{
			Purl:  "pkg:npm/%40angular/elements",
			Limit: 2,
		},
	}

	for _, dtoCompVersionInput := range goodTable {
		versions, err := compUc.GetComponentVersions(dtoCompVersionInput)
		if err != nil {
			t.Fatalf("an error '%s' was not expected when getting components", err)
		}
		fmt.Printf("Versions response: %+v\n", versions)
	}

	failTestTable := []dtos.ComponentVersionsInput{
		{
			Purl:  "",
			Limit: 0,
		},
		{
			Purl:  "pkg::NOEXIST::/%40angular/elements",
			Limit: 2,
		},
	}

	for _, dtoCompVersionInput := range failTestTable {
		_, err := compUc.GetComponentVersions(dtoCompVersionInput)
		if err == nil {
			t.Errorf("an error was expected when getting components version %v\n", err)
		}

	}
}
