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
	zlog "scanoss.com/components/pkg/logger"
	"testing"
)

func TestAllUrlsSearch(t *testing.T) {
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
	conn, err := db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	err = LoadTestSqlData(db, ctx, conn)
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	CloseConn(conn)

	allUrlsModel := NewAllUrlModel(ctx, db)
	allUrls, err := allUrlsModel.GetUrlsByPurlNameType("tablestyle", "gem", -1)
	if err != nil {
		t.Errorf("all_urls.GetUrlsByPurlName() error = %v", err)
	}
	if len(allUrls) == 0 {
		t.Errorf("all_urls.GetUrlsByPurlName() No URLs returned from query")
	}
	fmt.Printf("All Urls: %+v\n", allUrls)

	allUrls, err = allUrlsModel.GetUrlsByPurlNameType("NONEXISTENT", "none", -1)
	if err != nil {
		t.Errorf("all_urls.GetUrlsByPurlName() error = %v", err)
	}
	if len(allUrls) > 0 {
		t.Errorf("all_urls.GetUrlsByPurlNameType() URLs found when none should be: %v", allUrlsModel)
	}
	fmt.Printf("No Urls: %+v\n", allUrls)

	allUrls, err = allUrlsModel.GetUrlsByPurlNameType("", "none", -1)
	if err == nil {
		t.Errorf("An error was expected with empty purlName all_urls.GetUrlsByPurlName() error = %v", err)
	}

	allUrls, err = allUrlsModel.GetUrlsByPurlNameType("pkg:gem/tablestyle", "", -1)
	if err == nil {
		t.Errorf("An error was expected with empty purlType all_urls.GetUrlsByPurlName() error = %v", err)
	}

	allUrls, err = allUrlsModel.GetUrlsByPurlString("pkg:gem/tablestyle", -1)
	if err != nil {
		t.Errorf("all_urls.GetUrlsByPurlString() error = %v", err)
	}
	if len(allUrls) == 0 {
		t.Errorf("all_urls.GetUrlsByPurlString() No URLs returned from query")
	}
	fmt.Printf("All Urls: %+v\n", allUrls)

	allUrls, err = allUrlsModel.GetUrlsByPurlString("", -1)
	if err == nil {
		t.Errorf("An error was expected with empty purlString all_urls.GetUrlsByPurlString() error = %v", err)
	}

	allUrls, err = allUrlsModel.GetUrlsByPurlString("pkg::pypi", -1)
	if err == nil {
		t.Errorf("An error was expected with broken purlString all_urls.GetUrlsByPurlString() error = %v", err)
	}

}
