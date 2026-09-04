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

	"github.com/google/go-cmp/cmp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	_ "modernc.org/sqlite"
	myconfig "scanoss.com/components/pkg/config"
)

func TestDbLoad(t *testing.T) {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	db.SetMaxOpenConns(1)

	err = zlog.NewSugaredDevLogger()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer CloseDB(db)
	// The fixtures carry data only, so the schema has to exist first.
	if err = LoadTestSchema(db, nil, nil); err != nil {
		t.Fatalf("failed to load the test schema: %v", err)
	}
	err = loadSQLData(db, nil, nil, "./tests/component.sql")
	if err != nil {
		t.Errorf("failed to load SQL test data: %v", err)
	}
	err = LoadTestSQLData(db, nil, nil)
	if err != nil {
		t.Errorf("failed to load SQL test data: %v", err)
	}
	err = loadSQLData(db, nil, nil, "./tests/does-not-exist.sql")
	if err == nil {
		t.Errorf("did not fail to load SQL test data")
	}
	err = loadTestSQLDataFiles(db, nil, nil, []string{"./tests/does-not-exist.sql"})
	if err == nil {
		t.Errorf("did not fail to load SQL test data")
	}
	err = loadSQLData(db, nil, nil, "./tests/bad_sql.sql")
	if err == nil {
		t.Errorf("did not fail to load SQL test data")
	}
}

func TestRunQueriesInParallel(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	db := sqliteSetup(t) // Setup SQL Lite DB
	defer CloseDB(db)
	err = LoadTestSQLData(db, ctx, nil)
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	myConfig, err := myconfig.NewServerConfig(nil)
	if err != nil {
		t.Fatalf("failed to load Config: %v", err)
	}
	myConfig.Database.Trace = true
	q := database.NewDBSelectContext(s, db, nil, myConfig.Database.Trace)

	queryJobs := []QueryJob{
		{
			Query: db.Rebind("SELECT c.purl AS purl, c.component AS component" +
				" FROM component c LEFT JOIN component_metrics cm ON cm.purl_id = c.purl_id" +
				" WHERE c.component LIKE ?" +
				" LIMIT ?"),
			Args: []any{"%angular%", 2},
		},
	}
	res, err := RunQueries[Component](q, ctx, queryJobs)
	if err != nil {
		t.Errorf("Error running multiple queries %v", err)
	}
	fmt.Printf("Result of running queries %v:\n%v\n ", queryJobs, res)
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
		URL:       "",
	}

	comp1 := Component{
		Component: "scanner",
		PurlType:  "npm",
		PurlName:  "scanner",
		URL:       "https://www.npmjs.com/package/scanner",
	}

	comp1Similar := Component{
		Component: "scanner",
		PurlType:  "npm",
		PurlName:  "scanner",
		URL:       "www.npmjs.com/package/scanner",
	}

	comp2 := Component{
		Component: "graph",
		PurlType:  "npm",
		PurlName:  "graph",
		URL:       "https://www.npmjs.com/package/graph",
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
		if result := RemoveDuplicated[Component](test.input); !cmp.Equal(result, test.want) {
			diff := cmp.Diff(result, test.want)
			t.Fatalf("Expected %v and got %v\n Differences: %v", result, test.want, diff)
		}
	}
}
