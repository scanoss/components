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
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"scanoss.com/components/pkg/dtos"
	zlog "scanoss.com/components/pkg/logger"
	"scanoss.com/components/pkg/models"
	"testing"
)

func TestSearchComponentsUseCase(t *testing.T) {
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
	conn, err := db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	err = models.LoadTestSqlData(db, ctx, conn)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when loading test data", err)
	}
	models.CloseConn(conn)

	compUc := NewComponents(ctx, db)

	var compRequestData = `{
		"component": "angular",
		"package": "github"
	}
	`
	requestDto, err := dtos.ParseComponentSearchInput([]byte(compRequestData))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when parsing input json", err)
	}
	components, err := compUc.SearchComponents(requestDto)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when getting components", err)
	}
	fmt.Printf("Components response: %+v\n", components)

}

func TestGetComponentVersionsUseCase(t *testing.T) {
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
	conn, err := db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	err = models.LoadTestSqlData(db, ctx, conn)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when loading test data", err)
	}
	models.CloseConn(conn)

	var compVersionRequestData = `{
		"purl": "pkg:gem/tablestyle",
		"limit": 10
	}
	`
	compUc := NewComponents(ctx, db)
	requestDto, err := dtos.ParseComponentVersionsInput([]byte(compVersionRequestData))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when parsing input json", err)
	}
	versions, err := compUc.GetComponentVersions(requestDto)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when getting components", err)
	}
	fmt.Printf("Versions response: %+v\n", versions)

}
