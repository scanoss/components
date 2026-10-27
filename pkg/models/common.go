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

// This file common tasks for the models package

package models

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

// loadSQLData Load the specified SQL files into the supplied DB.
func loadSQLData(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn, filename string) error {
	fmt.Printf("Loading test data file: %v\n", filename)
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if conn != nil {
		_, err = conn.ExecContext(ctx, string(file))
	} else {
		_, err = db.Exec(string(file))
	}
	if err != nil {
		return err
	}
	return nil
}

// testDataFiles are the data fixtures loaded on top of the schema. They carry data
// only; the tables come from testSchemaDDL in test_schema.go.
var testDataFiles = []string{
	"../models/tests/component.sql",
	"../models/tests/component_metrics.sql",
	"../models/tests/component_status.sql",
	"../models/tests/component_version.sql",
	"../models/tests/component_version_status.sql",
	"../models/tests/component_version_license.sql",
	"../models/tests/license.sql",
	"../models/tests/spdx_license_data.sql",
	"../models/tests/db_version.sql",
}

// idempotentDDL rewrites CREATE TABLE/INDEX into their IF NOT EXISTS form so a test
// can load the schema more than once. testSchemaDDL transcribes production and must
// not carry test-only clauses, so the rewrite happens here instead of in the constant.
func idempotentDDL(sql string) string {
	for _, kind := range []string{"TABLE", "INDEX"} {
		bare := "CREATE " + kind + " "
		guarded := bare + "IF NOT EXISTS "
		// Normalise first, so an already-guarded statement is not double-guarded.
		sql = strings.ReplaceAll(sql, guarded, bare)
		sql = strings.ReplaceAll(sql, bare, guarded)
	}
	return sql
}

// LoadTestSchema creates the serving schema in the supplied DB. Call this before
// loading any data fixture, since the fixtures do not define their own tables.
// It is safe to call more than once on the same DB.
func LoadTestSchema(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn) error {
	var err error
	if conn != nil {
		_, err = conn.ExecContext(ctx, idempotentDDL(testSchemaDDL))
	} else {
		_, err = db.Exec(idempotentDDL(testSchemaDDL))
	}
	if err != nil {
		return fmt.Errorf("failed to load the test schema: %w", err)
	}
	return nil
}

// LoadTestSQLData loads the serving schema plus all the test data fixtures.
func LoadTestSQLData(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn) error {
	// In-memory SQLite gives each pooled connection its OWN empty database, and the
	// search runs its passes concurrently, so a second connection would find no
	// tables. Pin the pool to one so every query sees the same database.
	db.SetMaxOpenConns(1)
	if err := LoadTestSchema(db, ctx, conn); err != nil {
		return err
	}
	return loadTestSQLDataFiles(db, ctx, conn, testDataFiles)
}

// loadTestSQLDataFiles loads a list of test SQL files.
func loadTestSQLDataFiles(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn, files []string) error {
	for _, file := range files {
		err := loadSQLData(db, ctx, conn, file)
		if err != nil {
			return err
		}
	}
	return nil
}

// sqliteSetup sets up an in-memory SQL Lite DB for testing.
func sqliteSetup(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return db
}

// sqliteConn sets up a connection to a test DB.
func sqliteConn(t *testing.T, ctx context.Context, db *sqlx.DB) *sqlx.Conn {
	conn, err := db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return conn
}

// CloseDB closes the specified DB and logs any errors.
func CloseDB(db *sqlx.DB) {
	if db != nil {
		err := db.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing DB: %v", err)
		}
	}
}

// CloseConn closes the specified DB connection and logs any errors.
func CloseConn(conn *sqlx.Conn) {
	if conn != nil {
		err := conn.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing DB connection: %v", err)
		}
	}
}

type QueryJob struct {
	Query string
	Args  []any
}

type job struct {
	jobID int
	query string
	args  []any
}

type result[T any] struct {
	jobID int
	query string
	err   error
	dest  []T
}

func workerQuery[T any](q *database.DBQueryContext, ctx context.Context, jobs chan job, results chan result[T]) {
	var structResults []T
	for j := range jobs {
		err := q.SelectContext(ctx, &structResults, j.query, j.args...)
		results <- result[T]{
			jobID: j.jobID,
			query: j.query,
			err:   err,
			dest:  structResults,
		}
	}
}

func RunQueries[T any](q *database.DBQueryContext, ctx context.Context, queryJobs []QueryJob) ([]T, error) {
	numJobs := len(queryJobs)
	jobChan := make(chan job, numJobs)
	resultChan := make(chan result[T], numJobs)

	for w := 1; w <= numJobs; w++ {
		go workerQuery(q, ctx, jobChan, resultChan)
	}

	for i, queryJob := range queryJobs {
		jobChan <- job{
			jobID: i,
			query: queryJob.Query,
			args:  queryJob.Args,
		}
	}
	close(jobChan)

	resMap := make(map[int][]T)
	for i := 0; i < numJobs; i++ {
		res := <-resultChan
		if res.err == nil {
			resMap[res.jobID] = res.dest
		} else {
			return []T{}, res.err
		}
	}

	var output []T
	for i := 0; i <= numJobs; i++ {
		if v, ok := resMap[i]; ok {
			output = append(output, v...)
		}
	}
	return output, nil
}

func RemoveDuplicated[T comparable](items []T) []T {
	var unique []T
	m := make(map[T]int)
	for _, item := range items {
		if _, exist := m[item]; !exist {
			m[item] = len(unique)
			unique = append(unique, item)
		}
	}
	return unique
}
