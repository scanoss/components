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
	"strings"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	_ "modernc.org/sqlite"
	myconfig "scanoss.com/components/pkg/config"
)

//goland:noinspection DuplicatedCode
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
	component := NewComponentModel(ctx, s, database.NewDBSelectContext(s, db, conn, myConfig.Database.Trace), db, database.GetLikeOperator(db))

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

// The #ORDER placeholder is gone: the ranking clause is chosen per purl type and
// written straight into the query, so what is worth pinning is the choice itself.
func TestOrderByPurlType(t *testing.T) {
	tests := []struct {
		purlType string
		wantCols string
	}{
		{"github", "cm.git_created_at"},
		{"npm", "cm.first_version_date"},
		{"pypi", "cm.first_version_date"},
		{"gem", "cm.first_version_date"},
		// A purl type with no ranking entry gets no ORDER BY at all, as in mining.
		{"NONEXISTENT", ""},
	}
	for _, tt := range tests {
		t.Run(tt.purlType, func(t *testing.T) {
			got := orderByPurlType[tt.purlType]
			if tt.wantCols == "" {
				if got != "" {
					t.Errorf("order clause = %q, want none", got)
				}
				return
			}
			if !strings.Contains(got, tt.wantCols) {
				t.Errorf("order clause = %q, want it to rank on %s", got, tt.wantCols)
			}
			// The columns must be qualified for the metrics alias, or the join is pointless.
			if !strings.HasPrefix(got, "ORDER BY cm.") {
				t.Errorf("order clause = %q, want it qualified with the component_metrics alias", got)
			}
		})
	}
}

// The purl type used to be a `mines.purl_type = ?` filter; serving folds it into
// a prefix on the purl itself.
func TestPurlPatterns(t *testing.T) {
	if got := typePrefix("github"); got != "pkg:github/%" {
		t.Errorf("typePrefix = %q, want pkg:github/%%", got)
	}
	if got := namePattern("npm", "%lodash%"); got != "pkg:npm/%lodash%" {
		t.Errorf("namePattern = %q, want pkg:npm/%%lodash%%", got)
	}
}

func TestSplitPurl(t *testing.T) {
	tests := []struct {
		purl     string
		wantType string
		wantName string
	}{
		{"pkg:github/scanoss/engine", "github", "scanoss/engine"},
		{"pkg:npm/lodash", "npm", "lodash"},
		// Scoped npm packages stay percent-encoded; splitting must not touch them.
		{"pkg:npm/%40scope/name", "npm", "%40scope/name"},
		{"pkg:gem", "gem", ""},
	}
	for _, tt := range tests {
		t.Run(tt.purl, func(t *testing.T) {
			gotType, gotName := splitPurl(tt.purl)
			if gotType != tt.wantType || gotName != tt.wantName {
				t.Errorf("splitPurl(%q) = (%q, %q), want (%q, %q)", tt.purl, gotType, gotName, tt.wantType, tt.wantName)
			}
		})
	}
}
