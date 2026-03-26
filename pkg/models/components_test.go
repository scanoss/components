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
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	_ "modernc.org/sqlite"
	myconfig "scanoss.com/components/pkg/config"
	"testing"
)

func TestComponentsModel(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	db := sqliteSetup(t) // Setup SQL Lite DB
	defer CloseDB(db)
	conn := sqliteConn(t, ctx, db) // Get a connection from the pool
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
	db.DriverName()
	component := NewComponentModel(ctx, s, database.NewDBSelectContext(s, db, conn, myConfig.Database.Trace), database.GetLikeOperator(db))

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

		components, err = component.GetComponentsByNameVendorType(test.SearchParam, test.SearchParam, test.PurlType, test.Limit, test.Offset)
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

	_, err = component.GetComponentsByVendorType("", "", 0, 0)
	if err == nil {
		t.Errorf("An error was expected")
	}

	_, err = component.GetComponentsByNameVendorType("", "", "", 0, 0)
	if err == nil {
		t.Errorf("An error was expected")
	}
}

func TestPreProcessQueryJobs(t *testing.T) {

	testTable := []struct {
		qList    []QueryJob
		purlType string
		wanted   []QueryJob
	}{
		{
			qList: []QueryJob{
				{
					Query: "SELECT * project #ORDER LIMIT $1",
					Args:  []any{0},
				},
				{
					Query: "SELECT * project #ORDER LIMIT $1",
					Args:  []any{0},
				},
			},
			purlType: "github",
			wanted: []QueryJob{
				{
					Query: "SELECT * project ORDER BY git_created_at NULLS LAST , git_forks DESC NULLS LAST, git_stars DESC NULLS LAST LIMIT $1",
					Args:  []any{0},
				},
				{
					Query: "SELECT * project ORDER BY git_created_at NULLS LAST , git_forks DESC NULLS LAST, git_stars DESC NULLS LAST LIMIT $1",
					Args:  []any{0},
				},
			},
		},
		{
			qList: []QueryJob{{
				Query: "SELECT * FROM project #ORDER",
			}},
			purlType: "NONEXISTENT",
			wanted: []QueryJob{
				{
					Query: "SELECT * FROM project",
				},
			},
		},
		{
			qList: []QueryJob{
				{Query: "SELECT * FROM project"},
			},
			purlType: "github",
			wanted: []QueryJob{
				{Query: "SELECT * FROM project"},
			},
		},
	}

	for _, testInput := range testTable {
		q, err := preProcessQueryJob(testInput.qList, testInput.purlType)
		if err != nil {
			t.Errorf("Error produced when pre processing QueryJobs: %v\n", err)
		}
		if !cmp.Equal(q, testInput.wanted) {
			t.Errorf(" unexpected output for preProcessQueryJob()\nWanted: %v\n, Got:%v\n", testInput.wanted, q)
		}
	}

	_, err := preProcessQueryJob([]QueryJob{}, "")
	if err == nil {
		t.Errorf("An error was expected with empty parameters \n")
	}
}
