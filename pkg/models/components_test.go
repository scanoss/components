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

func TestComponentsSearch(t *testing.T) {
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
	defer CloseDB(db)
	conn, err := db.Connx(ctx) // Get a connection from the pool (with context)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer CloseConn(conn)
	err = loadTestSqlDataFiles(db, ctx, conn, []string{"../models/tests/projects.sql", "../models/tests/mines.sql"})
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}

	component := NewComponentModel(ctx, conn)
	var purlType = "npm"
	var compName = "uuid"
	fmt.Printf("Searching for components: Component Name:%v, PurlType: %v\n", compName, purlType)
	components, err := component.GetComponentsByNameType(compName, purlType, -1, -1)
	if err != nil {
		t.Errorf("components.GetComponentsByNameType() error = %v", err)
	}
	if len(components) < 1 {
		t.Errorf("components.GetComponentsByNameType() No components returned from query")
	}
	fmt.Printf("Components: %v\n", components)

	purlType = "github"
	compName = "angular"
	fmt.Printf("Searching for components: Component Name:%v, PurlType: %v\n", compName, purlType)
	components, err = component.GetComponentsByNameType(compName, purlType, 20, -1)
	if err != nil {
		t.Errorf("components.GetComponentsByNameType() error = %v", err)
	}
	if len(components) < 1 {
		t.Errorf("components.GetComponentsByNameType() No components returned from query")
	}
	if len(components) > 20 {
		t.Errorf("components.GetComponentsByNameType() Limit of components retrieved exedeed the maximum")
	}

	fmt.Printf("Components: %v\n", components)

}
