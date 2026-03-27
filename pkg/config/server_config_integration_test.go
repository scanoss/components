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

package config

import (
	"fmt"
	"os"
	"testing"

	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

// TestServerConfig_StatusMapping_FromEnv verifies that custom status mappings can be loaded from environment variables.
// Tests that STATUS_MAPPING env var is correctly parsed as JSON and applied to the StatusMapper.
// Verifies both custom mappings and default fallback behavior for non-overridden keys.
func TestServerConfig_StatusMapping_FromEnv(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("Failed to initialise logger: %v", err)
	}
	defer zlog.SyncZap()
	envValue := `{"unlisted":"custom-removed","yanked":"custom-yanked"}`
	errEnv := os.Setenv("STATUS_MAPPING", envValue)
	if errEnv != nil {
		t.Fatalf("Could not set env variable: %v", errEnv)
	}
	cfg, err := NewServerConfig(nil)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if ue := os.Unsetenv("STATUS_MAPPING"); ue != nil {
		fmt.Printf("Warning: Problem running Unsetenv: %v\n", ue)
	}
	// Allowing the GetStatusMapper to load the config
	mapper := cfg.GetStatusMapper()
	if mapper == nil {
		t.Fatal("Expected non-nil mapper from GetStatusMapper()")
	}
	result := mapper.MapStatus("unlisted")
	if result != "custom-removed" {
		t.Errorf("Expected 'custom-removed', got %q", result)
	}
	result = mapper.MapStatus("yanked")
	if result != "custom-yanked" {
		t.Errorf("Expected 'custom-yanked', got %q", result)
	}
	result = mapper.MapStatus("deleted")
	if result != "deleted" {
		t.Errorf("Expected 'deleted', got %q", result)
	}
}

// TestServerConfig_StatusMapping_DefaultWhenNotSet verifies that default status mappings are used when STATUS_MAPPING is not configured.
// Tests that StatusMapper initializes correctly with built-in default mappings.
// Ensures default behavior when no custom configuration is provided.
func TestServerConfig_StatusMapping_DefaultWhenNotSet(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("Failed to initialise logger: %v", err)
	}
	defer zlog.SyncZap()
	errEnv := os.Unsetenv("STATUS_MAPPING")
	if errEnv != nil {
		t.Fatalf("Could not set env variable: %v", errEnv)
	}
	cfg, err := NewServerConfig(nil)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	mapper := cfg.GetStatusMapper()
	if mapper == nil {
		t.Fatal("Expected non-nil mapper from GetStatusMapper()")
	}
	result := mapper.MapStatus("unlisted")
	if result != "removed" {
		t.Errorf("Expected 'removed', got %q", result)
	}
	result = mapper.MapStatus("yanked")
	if result != "removed" {
		t.Errorf("Expected 'removed', got %q", result)
	}
	result = mapper.MapStatus("active")
	if result != "active" {
		t.Errorf("Expected 'active', got %q", result)
	}
}

// TestServerConfig_StatusMapping_WithProvidedLogger verifies that StatusMapper receives and uses an explicitly provided logger.
// Simulates production usage where logger is initialized before config loading (two-phase initialization).
// Tests that custom mappings work correctly when passing an initialized logger to NewServerConfig.
func TestServerConfig_StatusMapping_WithProvidedLogger(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("Failed to initialise logger: %v", err)
	}
	defer zlog.SyncZap()
	envValue := `{"test-status":"test-mapped"}`
	errEnv := os.Setenv("STATUS_MAPPING", envValue)
	if errEnv != nil {
		t.Fatalf("Could not set env variable: %v", errEnv)
	}
	cfg, err := NewServerConfig(nil)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if ue := os.Unsetenv("STATUS_MAPPING"); ue != nil {
		fmt.Printf("Warning: Problem running Unsetenv: %v\n", ue)
	}

	cfg.InitStatusMapperConfig(zlog.S)
	mapper := cfg.GetStatusMapper()
	if mapper == nil {
		t.Fatal("Expected non-nil mapper from GetStatusMapper()")
	}
	result := mapper.MapStatus("test-status")
	if result != "test-mapped" {
		t.Errorf("Expected 'test-mapped', got %q", result)
	}
}
