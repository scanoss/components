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
	"os"
	"testing"

	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

func TestServerConfig_StatusMapping_FromEnv(t *testing.T) {
	// Initialize logger
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zlog.SyncZap()

	// Set environment variable
	envValue := `{"unlisted":"custom-removed","yanked":"custom-yanked"}`
	os.Setenv("STATUS_MAPPING", envValue)
	defer os.Unsetenv("STATUS_MAPPING")

	// Load config
	cfg, err := NewServerConfig(nil)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify status mapper was initialized
	if cfg.statusMapper == nil {
		t.Fatal("Expected statusMapper to be initialized")
	}

	// Test that custom mappings work
	mapper := cfg.GetStatusMapper()
	if mapper == nil {
		t.Fatal("Expected non-nil mapper from GetStatusMapper()")
	}

	// Test custom mapping
	result := mapper.MapStatus("unlisted")
	if result != "custom-removed" {
		t.Errorf("Expected 'custom-removed', got %q", result)
	}

	result = mapper.MapStatus("yanked")
	if result != "custom-yanked" {
		t.Errorf("Expected 'custom-yanked', got %q", result)
	}

	// Test default mapping still works for non-overridden keys
	result = mapper.MapStatus("deleted")
	if result != "deleted" {
		t.Errorf("Expected 'deleted', got %q", result)
	}
}

func TestServerConfig_StatusMapping_DefaultWhenNotSet(t *testing.T) {
	// Initialize logger
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zlog.SyncZap()

	// Make sure STATUS_MAPPING is not set
	os.Unsetenv("STATUS_MAPPING")

	// Load config
	cfg, err := NewServerConfig(nil)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify status mapper was initialized with defaults
	mapper := cfg.GetStatusMapper()
	if mapper == nil {
		t.Fatal("Expected non-nil mapper from GetStatusMapper()")
	}

	// Test default mappings
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
