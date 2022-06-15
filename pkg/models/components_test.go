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

	err = LoadTestSqlData(db, ctx, conn)
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	CloseConn(conn)

	component := NewComponentModel(ctx, db)
	var purlType = "npm"
	var compName = "react"
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
	components, err = component.GetComponentsByNameType(compName, purlType, 20, 0)
	if err != nil {
		t.Errorf("components.GetComponentsByNameType() error = %v", err)
	}
	fmt.Printf("Components: %v\n", components)

	purlType = "github"
	compName = "angular"
	fmt.Printf("Searching for components: Component Name:%v, PurlType: %v\n", compName, purlType)
	components, err = component.GetComponents(compName, purlType, 4, 0)
	if err != nil {
		t.Errorf("components.GetComponentsByNameType() error = %v", err)
	}
	fmt.Printf("Components: %v\n", components)
}

func TestRemoveDuplicates(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}

	var components = []Component{
		{
			Component: "comp1",
			PurlType:  "purlType",
			PurlName:  "purlName",
			Url:       "url",
		},
		{
			Component: "angular",
			PurlType:  "github",
			PurlName:  "angular",
			Url:       "url",
		},
		{
			Component: "comp1",
			PurlType:  "purlType",
			PurlName:  "purlName",
			Url:       "url",
		},
	}

	var uniqueComponents []Component
	uniqueComponents = removeDuplicateComponents(components)

	if len(uniqueComponents) != 2 {
		t.Fatalf("Expected only 2 elements")
	}

	fmt.Printf("%v+", uniqueComponents)

}
