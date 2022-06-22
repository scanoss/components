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
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	zlog "scanoss.com/components/pkg/logger"
)

var DEFAULT_MAX_VERSION_LIMIT = 50
var DEFAULT_MAX_COMPONENT_LIMIT = 50

// loadSqlData Load the specified SQL files into the supplied DB
func loadSqlData(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn, filename string) error {
	fmt.Printf("Loading test data file: %v\n", filename)
	file, err := ioutil.ReadFile(filename)
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

// LoadTestSqlData loads all the required test SQL files
func LoadTestSqlData(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn) error {
	files := []string{"../models/tests/mines.sql", "../models/tests/all_urls.sql", "../models/tests/projects.sql",
		"../models/tests/licenses.sql", "../models/tests/versions.sql"}
	return loadTestSqlDataFiles(db, ctx, conn, files)
}

// loadTestSqlDataFiles loads a list of test SQL files
func loadTestSqlDataFiles(db *sqlx.DB, ctx context.Context, conn *sqlx.Conn, files []string) error {
	for _, file := range files {
		err := loadSqlData(db, ctx, conn, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func CloseDB(db *sqlx.DB) {
	if db != nil {
		zlog.S.Debugf("Closing DB...")
		err := db.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing DB: %v", err)
		}
	}
}

func CloseConn(conn *sqlx.Conn) {
	if conn != nil {
		zlog.S.Debugf("Closing Connection...")
		err := conn.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing DB connection: %v", err)
		}
	}
}

func CloseRows(rows *sqlx.Rows) {
	if rows != nil {
		zlog.S.Debugf("Closing Rows...")
		err := rows.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing Rows: %v", err)
		}
	}
}

type QueryJob struct {
	query string
	args  []any
}

type job struct {
	jobId int
	query string
	args  []any
}

type result[T any] struct {
	jobId int
	query string
	err   error
	dest  []T
}

func workerQuery[T any](db *sqlx.DB, ctx context.Context, jobs chan job, results chan result[T]) {
	var structResults []T
	for j := range jobs {
		err := db.SelectContext(ctx, &structResults, j.query, j.args...)
		results <- result[T]{
			jobId: j.jobId,
			query: j.query,
			err:   err,
			dest:  structResults,
		}
	}
}

func RunQueriesInParallel[T any](db *sqlx.DB, ctx context.Context, queryJobs []QueryJob) ([]T, error) {
	numJobs := len(queryJobs)
	jobChan := make(chan job, numJobs)
	resultChan := make(chan result[T], numJobs)

	for w := 1; w <= numJobs; w++ {
		go workerQuery(db, ctx, jobChan, resultChan)
	}

	for i, queryJob := range queryJobs {
		jobChan <- job{
			jobId: i,
			query: queryJob.query,
			args:  queryJob.args,
		}
	}
	close(jobChan)

	resMap := make(map[int][]T)
	for i := 0; i < numJobs; i++ {
		res := <-resultChan
		if res.err == nil {
			resMap[res.jobId] = res.dest
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
