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
	"github.com/google/go-cmp/cmp"
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
	fmt.Printf("Searching for components by differents criteria: Component Name:%v, PurlType: %v\n", compName, purlType)
	components, err = component.GetComponents(compName, purlType, 20, 0)
	if err != nil {
		t.Errorf("components.GetComponentsByNameType() error = %v", err)
	}
	fmt.Printf("Components: %v\n", components)
}

func TestRemoveDuplicates(t *testing.T) {

	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a sugared logger", err)
	}
	compEmpty := Component{}

	compSemiComplete := Component{
		Component: "hyx-decrypt",
		PurlType:  "npm",
		PurlName:  "hyx-decrypt",
		Url:       "",
	}

	comp1 := Component{
		Component: "scanner",
		PurlType:  "npm",
		PurlName:  "scanner",
		Url:       "https://www.npmjs.com/package/scanner",
	}

	comp1Similar := Component{
		Component: "scanner",
		PurlType:  "npm",
		PurlName:  "scanner",
		Url:       "www.npmjs.com/package/scanner",
	}

	comp2 := Component{
		Component: "graph",
		PurlType:  "npm",
		PurlName:  "graph",
		Url:       "https://www.npmjs.com/package/graph",
	}

	testTable := []struct {
		input []Component
		want  []Component
	}{
		{input: []Component{comp1, comp1, comp1, comp1}, want: []Component{comp1}},
		{input: []Component{comp1, compEmpty, comp1, compSemiComplete}, want: []Component{comp1, compEmpty, compSemiComplete}},
		{input: []Component{compEmpty, compEmpty}, want: []Component{compEmpty}},
		{input: []Component{comp1, compEmpty, comp2}, want: []Component{comp1, compEmpty, comp2}},
		{input: []Component{comp1, comp1Similar}, want: []Component{comp1, comp1Similar}},
	}

	for _, test := range testTable {
		if result := removeDuplicateComponents(test.input); !cmp.Equal(result, test.want) {
			diff := cmp.Diff(result, test.want)
			t.Fatalf("Expected %v and got %v\n Differences: %v", result, test.want, diff)
		}
	}
}
