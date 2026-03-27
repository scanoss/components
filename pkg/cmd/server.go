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

package cmd

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golobby/config/v3"
	"github.com/golobby/config/v3/pkg/feeder"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/scanoss/go-grpc-helper/pkg/files"
	gd "github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	gs "github.com/scanoss/go-grpc-helper/pkg/grpc/server"
	gomodels "github.com/scanoss/go-models/pkg/models"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	_ "modernc.org/sqlite"
	myconfig "scanoss.com/components/pkg/config"
	"scanoss.com/components/pkg/protocol/grpc"
	"scanoss.com/components/pkg/protocol/rest"
	"scanoss.com/components/pkg/service"
)

//TODO: Now the config includes the app version.
//  This might be worth moving to the file pkg/config/server_config.go

//go:generate bash ../../get_version.sh
//go:embed version.txt
var version string

// getConfig checks command line args for option to feed into the config parser.
// It performs a two-phase initialization: first loads basic config to get logging settings,
// then initializes the logger and reloads config with the proper logger for StatusMapper.
func getConfig() (*myconfig.ServerConfig, error) {
	var jsonConfig, envConfig string
	flag.StringVar(&jsonConfig, "json-config", "", "Application JSON config")
	flag.StringVar(&envConfig, "env-config", "", "Application dot-ENV config")
	debug := flag.Bool("debug", false, "Enable debug")
	ver := flag.Bool("version", false, "Display current version")
	flag.Parse()
	if *ver {
		fmt.Printf("Version: %v", version)
		os.Exit(1)
	}
	var feeders []config.Feeder
	if len(jsonConfig) > 0 {
		feeders = append(feeders, feeder.Json{Path: jsonConfig})
	}
	if len(envConfig) > 0 {
		feeders = append(feeders, feeder.DotEnv{Path: envConfig})
	}
	if *debug {
		err := os.Setenv("APP_DEBUG", "1")
		if err != nil {
			fmt.Printf("Warning: Failed to set env APP_DEBUG to 1: %v", err)
			return nil, err
		}
	}
	myConfig, err := myconfig.NewServerConfig(feeders)
	if err != nil {
		return nil, err
	}
	// Initialize the application logger
	err = zlog.SetupAppLogger(myConfig.App.Mode, myConfig.Logging.ConfigFile, myConfig.App.Debug)
	if err != nil {
		return nil, err
	}
	// Initialise the status mapping config
	myConfig.InitStatusMapperConfig(zlog.S)
	return myConfig, err
}

// RunServer runs the gRPC Component Server.
func RunServer() error {
	// Load command line options and config (logger is initialized inside getConfig)
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	defer zlog.SyncZap()
	// Check if TLS/SSL should be enabled
	startTLS, err := files.CheckTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
	if err != nil {
		return err
	}
	// Check if IP filtering should be enabled
	allowedIPs, deniedIPs, err := files.LoadFiltering(cfg.Filtering.AllowListFile, cfg.Filtering.DenyListFile)
	if err != nil {
		return err
	}
	// Set the default version from the embedded binary version if not overridden by config/env
	if len(cfg.App.Version) == 0 {
		cfg.App.Version = strings.TrimSpace(version)
	}
	zlog.S.Infof("Starting SCANOSS Component Service: %v", cfg.App.Version)
	// Set up the database connection pool
	db, err := gd.OpenDBConnection(cfg.Database.Dsn, cfg.Database.Driver, cfg.Database.User, cfg.Database.Passwd,
		cfg.Database.Host, cfg.Database.Schema, cfg.Database.SslMode)
	if err != nil {
		return err
	}
	if err = gd.SetDBOptionsAndPing(db); err != nil {
		return err
	}
	defer gd.CloseDBConnection(db)
	// Log database version info
	logDBVersion(db)
	// Set up dynamic logging (if necessary)
	zlog.SetupAppDynamicLogging(cfg.Logging.DynamicPort, cfg.Logging.DynamicLogging)
	// Register the component service
	v2API := service.NewComponentServer(db, cfg)
	ctx := context.Background()
	// Start the REST grpc-gateway if requested
	var srv *http.Server
	if len(cfg.App.RESTPort) > 0 {
		if srv, err = rest.RunServer(cfg, ctx, cfg.App.GRPCPort, cfg.App.RESTPort, allowedIPs, deniedIPs, startTLS); err != nil {
			return err
		}
	}
	// Start the gRPC service
	server, err := grpc.RunServer(cfg, v2API, cfg.App.GRPCPort, allowedIPs, deniedIPs, startTLS)
	if err != nil {
		return err
	}
	// graceful shutdown
	return gs.WaitServerComplete(srv, server)
}

// logDBVersion logs the current version of the database.
func logDBVersion(db *sqlx.DB) {
	// Log database version info
	dbVersionModel := gomodels.NewDBVersionModel(db)
	dbVersion, dbVersionErr := dbVersionModel.GetCurrentVersion(context.Background())
	switch {
	case dbVersionErr != nil:
		zlog.S.Warnf("Could not read db_version table: %v", dbVersionErr)
	case len(dbVersion.SchemaVersion) > 0:
		zlog.S.Infof("Loaded decoration DB: package=%s, schema=%s, created_at=%s",
			dbVersion.PackageName, dbVersion.SchemaVersion, dbVersion.CreatedAt)
	default:
		zlog.S.Warn("db_version table is empty")
	}
}
