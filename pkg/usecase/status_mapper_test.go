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

package usecase

import (
	"testing"

	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

func TestStatusMapper_MapStatus_DefaultMappings(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	s := zlog.S

	// Create mapper with empty JSON (should use defaults)
	mapper := NewStatusMapper(s, "")

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"active maps to active", "active", "active"},
		{"unlisted maps to removed", "unlisted", "removed"},
		{"yanked maps to removed", "yanked", "removed"},
		{"deleted maps to deleted", "deleted", "deleted"},
		{"deprecated maps to deprecated", "deprecated", "deprecated"},
		{"unpublished maps to removed", "unpublished", "removed"},
		{"archived maps to deprecated", "archived", "deprecated"},
		{"ACTIVE (uppercase) maps to active", "ACTIVE", "active"},
		{"Unlisted (mixed case) maps to removed", "Unlisted", "removed"},
		{"unknown status returns original", "unknown-status", "unknown-status"},
		{"empty string returns empty", "", ""},
		{"whitespace status returns original", "  some status  ", "  some status  "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapper.MapStatus(tc.input)
			if result != tc.expected {
				t.Errorf("MapStatus(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestStatusMapper_MapStatus_CustomMappings(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	s := zlog.S

	// Create mapper with custom JSON mappings
	customJSON := `{"unlisted":"custom-removed","new-status":"custom-value","active":"still-active"}`
	mapper := NewStatusMapper(s, customJSON)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"custom unlisted mapping", "unlisted", "custom-removed"},
		{"custom new-status mapping", "new-status", "custom-value"},
		{"custom active override", "active", "still-active"},
		{"default yanked still works", "yanked", "removed"},
		{"default deleted still works", "deleted", "deleted"},
		{"unknown status returns original", "completely-unknown", "completely-unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapper.MapStatus(tc.input)
			if result != tc.expected {
				t.Errorf("MapStatus(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestStatusMapper_MapStatus_InvalidJSON(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	s := zlog.S

	// Create mapper with invalid JSON (should fall back to defaults)
	invalidJSON := `{this is not valid json}`
	mapper := NewStatusMapper(s, invalidJSON)

	// Should use defaults when JSON is invalid
	result := mapper.MapStatus("unlisted")
	if result != "removed" {
		t.Errorf("MapStatus with invalid JSON should use defaults, got %q, expected %q", result, "removed")
	}

	result = mapper.MapStatus("active")
	if result != "active" {
		t.Errorf("MapStatus with invalid JSON should use defaults, got %q, expected %q", result, "active")
	}
}

func TestStatusMapper_MapStatus_EmptyJSON(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	s := zlog.S

	// Test various empty JSON scenarios
	emptyScenarios := []string{"", "   ", "{}", "   {}   "}

	for _, emptyJSON := range emptyScenarios {
		mapper := NewStatusMapper(s, emptyJSON)

		// Should use defaults
		result := mapper.MapStatus("unlisted")
		if result != "removed" {
			t.Errorf("MapStatus with empty JSON %q should use defaults, got %q, expected %q", emptyJSON, result, "removed")
		}
	}
}

func TestStatusMapper_MapStatus_CaseSensitivity(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	s := zlog.S

	mapper := NewStatusMapper(s, "")

	// Test that mapping is case-insensitive for lookup
	testCases := []struct {
		input    string
		expected string
	}{
		{"active", "active"},
		{"ACTIVE", "active"},
		{"Active", "active"},
		{"AcTiVe", "active"},
		{"unlisted", "removed"},
		{"UNLISTED", "removed"},
		{"Unlisted", "removed"},
		{"UnLiStEd", "removed"},
	}

	for _, tc := range testCases {
		result := mapper.MapStatus(tc.input)
		if result != tc.expected {
			t.Errorf("MapStatus(%q) = %q, expected %q (case-insensitive)", tc.input, result, tc.expected)
		}
	}
}

func TestGetDefaultStatusMapping(t *testing.T) {
	defaults := getDefaultStatusMapping()

	expectedMappings := map[string]string{
		"active":      "active",
		"unlisted":    "removed",
		"yanked":      "removed",
		"deleted":     "deleted",
		"deprecated":  "deprecated",
		"unpublished": "removed",
		"archived":    "deprecated",
	}

	if len(defaults) != len(expectedMappings) {
		t.Errorf("Expected %d default mappings, got %d", len(expectedMappings), len(defaults))
	}

	for key, expectedValue := range expectedMappings {
		if actualValue, exists := defaults[key]; !exists {
			t.Errorf("Default mapping missing key %q", key)
		} else if actualValue != expectedValue {
			t.Errorf("Default mapping for %q: got %q, expected %q", key, actualValue, expectedValue)
		}
	}
}
