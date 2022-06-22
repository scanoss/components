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

func TestDbLoad(t *testing.T) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer CloseDB(db)
	err = loadSqlData(db, nil, nil, "./tests/mines.sql")
	if err != nil {
		t.Errorf("failed to load SQL test data: %v", err)
	}
	err = LoadTestSqlData(db, nil, nil)
	if err != nil {
		t.Errorf("failed to load SQL test data: %v", err)
	}
	err = loadSqlData(db, nil, nil, "./tests/does-not-exist.sql")
	if err == nil {
		t.Errorf("did not fail to load SQL test data")
	}
	err = loadTestSqlDataFiles(db, nil, nil, []string{"./tests/does-not-exist.sql"})
	if err == nil {
		t.Errorf("did not fail to load SQL test data")
	}
	err = loadSqlData(db, nil, nil, "./tests/bad_sql.sql")
	if err == nil {
		t.Errorf("did not fail to load SQL test data")
	}
}

func TestRunQueriesInParallel(t *testing.T) {
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

	queryJobs := []QueryJob{
		{
			query: "SELECT component, purl_name, m.purl_type FROM projects p " +
				" LEFT JOIN mines m ON p.mine_id = m.id" +
				" WHERE p.component LIKE $1" +
				" LIMIT $2",
			args: []any{"%angular%", 2},
		},
	}
	res, err := RunQueriesInParallel[Component](db, ctx, queryJobs)
	if err != nil {
		t.Errorf("Error running multiple queries %v", err)
	}
	fmt.Printf("Result of running queries %v:\n%v\n ", queryJobs, res)

	queryJobs = []QueryJob{
		{
			query: "SELECT id, name FROM mines LIMIT 1",
		},
		{
			query: "SELECT purl_type FROM mines LIMIT $1",
			args:  []any{3},
		},
	}
	res1, err := RunQueriesInParallel[Mine](db, ctx, queryJobs)
	if err != nil {
		t.Errorf("Error running multiple queries %v", err)
	}
	fmt.Printf("Result of running queries %v:\n%v\n ", queryJobs, res1)
}
