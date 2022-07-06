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
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	zlog "scanoss.com/components/pkg/logger"
	"testing"
)

func TestComponentsModel(t *testing.T) {
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
	db.SetMaxOpenConns(1)
	defer CloseDB(db)
	conn, err := db.Connx(ctx) // Get a connection from the pool (with context)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	err = LoadTestSqlData(db, ctx, conn)
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	CloseConn(conn)

	component := NewComponentModel(ctx, db)

	passTestTable := []struct {
		SearchParam string
		PurlType    string
		Limit       int
		Offset      int
	}{
		{
			SearchParam: "angular",
		},
		{
			SearchParam: "angular",
			Limit:       100,
		},
		{
			SearchParam: "angular",
			PurlType:    "",
			Limit:       -10,
			Offset:      -10,
		},
	}

	for _, test := range passTestTable {
		fmt.Printf("Searching:%v, PurlType: %v, Limit: %v, Offset: %v\n",
			test.SearchParam,
			test.PurlType,
			test.Limit,
			test.Offset)

		components, err := component.GetComponents(test.SearchParam, test.PurlType, test.Limit, test.Offset)
		if err != nil {
			t.Errorf("components.GetComponents() error = %v", err)
		}
		fmt.Printf("Components: %v\n", components)

		components, err = component.GetComponentsByNameType(test.SearchParam, test.PurlType, test.Limit, test.Offset)
		if err != nil {
			t.Errorf("components.GetComponentsByNameType() error = %v", err)
		}
		fmt.Printf("Components: %v\n", components)

		components, err = component.GetComponentsByVendorType(test.SearchParam, test.PurlType, test.Limit, test.Offset)
		if err != nil {
			t.Errorf("components.GetComponentsByVendorType() error = %v", err)
		}
		fmt.Printf("Components: %v\n", components)
	}

	_, err = component.GetComponents("", "", 0, 0)
	if err == nil {
		t.Errorf("An error was expected")
	}

	_, err = component.GetComponentsByNameType("", "", 0, 0)
	if err == nil {
		t.Errorf("An error was expected")
	}

	_, err = component.GetComponentsByNameType("", "", 0, 0)
	if err == nil {
		t.Errorf("An error was expected")
	}

	_, err = component.GetComponentsByVendorType("", "", 0, 0)
	if err == nil {
		t.Errorf("An error was expected")
	}
}
