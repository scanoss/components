// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2026 SCANOSS.COM
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

package config

import (
	"encoding/json"

	"github.com/golobby/config/v3"
	"github.com/golobby/config/v3/pkg/feeder"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"go.uber.org/zap"
)

const (
	defaultGrpcPort = "50053"
	defaultRestPort = "40053"
)

// parseStatusMappingString converts a string to interface{} for StatusMapper
// It handles both JSON object format (from config file) and JSON string format (from env var).
func parseStatusMappingString(s string) interface{} {
	if s == "" {
		return nil
	}
	// Try to unmarshal as map first (JSON object from config file)
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err == nil {
		return m
	}
	// Otherwise return as string (JSON string from env var)
	return s
}

// ServerConfig is a configuration for Server.
type ServerConfig struct {
	App struct {
		Name           string `env:"APP_NAME"`
		Version        string `env:"APP_VERSION"`
		GRPCPort       string `env:"APP_PORT"`
		RESTPort       string `env:"REST_PORT"`
		Debug          bool   `env:"APP_DEBUG"`           // true/false
		Trace          bool   `env:"APP_TRACE"`           // true/false
		Mode           string `env:"APP_MODE"`            // dev or prod
		GRPCReflection bool   `env:"APP_GRPC_REFLECTION"` // Enables gRPC reflection service for debugging and discovery
	}
	Logging struct {
		DynamicLogging bool   `env:"LOG_DYNAMIC"`      // true/false
		DynamicPort    string `env:"LOG_DYNAMIC_PORT"` // host:port
		ConfigFile     string `env:"LOG_JSON_CONFIG"`
	}
	Telemetry struct {
		Enabled      bool   `env:"OTEL_ENABLED"`       // true/false
		OltpExporter string `env:"OTEL_EXPORTER_OLTP"` // OTEL OLTP exporter (default 0.0.0.0:4317)
	}
	Database struct {
		Driver  string `env:"DB_DRIVER"`
		Host    string `env:"DB_HOST"`
		User    string `env:"DB_USER"`
		Passwd  string `env:"DB_PASSWD"`
		Schema  string `env:"DB_SCHEMA"`
		SslMode string `env:"DB_SSL_MODE"` // enable/disable
		Dsn     string `env:"DB_DSN"`
		Trace   bool   `env:"DB_TRACE"` // true/false
	}
	TLS struct {
		CertFile string `env:"COMP_TLS_CERT"` // TLS Certificate
		KeyFile  string `env:"COMP_TLS_KEY"`  // Private TLS Key
		CN       string `env:"COMP_TLS_CN"`   // Common Name (replaces the CN on the certificate)
	}
	Filtering struct {
		AllowListFile  string `env:"COMP_ALLOW_LIST"`       // Allow list file for incoming connections
		DenyListFile   string `env:"COMP_DENY_LIST"`        // Deny list file for incoming connections
		BlockByDefault bool   `env:"COMP_BLOCK_BY_DEFAULT"` // Block request by default if they are not in the allow list
		TrustProxy     bool   `env:"COMP_TRUST_PROXY"`      // Trust the interim proxy or not (causes the source IP to be validated instead of the proxy)
	}
	StatusMapping struct {
		Mapping string `env:"STATUS_MAPPING"` // JSON string mapping DB statuses to classified statuses (from env or file)
	}
	// StatusMapper is the compiled status mapper (initialised once at startup)
	statusMapper *StatusMapper
}

// NewServerConfig loads all config options and return a struct for use.
func NewServerConfig(feeders []config.Feeder) (*ServerConfig, error) {
	cfg := ServerConfig{}
	setServerConfigDefaults(&cfg)
	c := config.New()
	for _, f := range feeders {
		c.AddFeeder(f)
	}
	c.AddFeeder(feeder.Env{})
	c.AddStruct(&cfg)
	err := c.Feed()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// setServerConfigDefaults attempts to set reasonable defaults for the server config.
func setServerConfigDefaults(cfg *ServerConfig) {
	cfg.App.Name = "SCANOSS Component Server"
	cfg.App.GRPCPort = defaultGrpcPort
	cfg.App.RESTPort = defaultRestPort
	cfg.App.Mode = "dev"
	cfg.App.GRPCReflection = false
	cfg.App.Debug = false
	cfg.Database.Driver = "postgres"
	cfg.Database.Host = "localhost"
	cfg.Database.User = "scanoss"
	cfg.Database.Schema = "scanoss"
	cfg.Database.SslMode = "disable"
	cfg.Database.Trace = false
	cfg.Logging.DynamicLogging = true
	cfg.Logging.DynamicPort = "localhost:60053"
	cfg.Telemetry.Enabled = false
	cfg.Telemetry.OltpExporter = "0.0.0.0:4317" // Default OTEL OLTP gRPC Exporter endpoint
}

// InitStatusMapperConfig initialise the status mapper for mapping component statuses
func (cfg *ServerConfig) InitStatusMapperConfig(s *zap.SugaredLogger) {
	cfg.statusMapper = NewStatusMapper(s, parseStatusMappingString(cfg.StatusMapping.Mapping))
}

// GetStatusMapper returns the status mapper for mapping database statuses to classified statuses.
func (cfg *ServerConfig) GetStatusMapper() *StatusMapper {
	// Initialise the mapper if it wasn't done previously
	if cfg.statusMapper == nil {
		cfg.InitStatusMapperConfig(zlog.S)
	}
	return cfg.statusMapper
}
