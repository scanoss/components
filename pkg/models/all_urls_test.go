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
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	myconfig "scanoss.com/components/pkg/config"
	"testing"
)

func TestAllUrlsSearch(t *testing.T) {
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

	allUrlsModel := NewAllUrlModel(ctx, s, database.NewDBSelectContext(s, db, conn, myConfig.Database.Trace))
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

	_, err = allUrlsModel.GetUrlsByPurlNameType("", "", 0)
	if err == nil {
		t.Errorf("An error was expected with empty purlName all_urls.GetUrlsByPurlName() error = %v", err)
	}

	_, err = allUrlsModel.GetUrlsByPurlNameType("pkg:gem/tablestyle", "", -1)
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

	_, err = allUrlsModel.GetUrlsByPurlString("", -1)
	if err == nil {
		t.Errorf("An error was expected with empty purlString all_urls.GetUrlsByPurlString() error = %v", err)
	}

	_, err = allUrlsModel.GetUrlsByPurlString("pkg::pypi", -1)
	if err == nil {
		t.Errorf("An error was expected with broken purlString all_urls.GetUrlsByPurlString() error = %v", err)
	}

}
