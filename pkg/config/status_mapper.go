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
	"strings"

	"go.uber.org/zap"
)

// StatusMapper handles mapping of database statuses to classified statuses
type StatusMapper struct {
	mapping map[string]string
	s       *zap.SugaredLogger
}

// NewStatusMapper creates a new StatusMapper with the provided mapping
// mappingConfig can be:
//   - map[string]interface{} (from JSON config file)
//   - string (from environment variable, containing JSON)
//   - nil or empty (uses default mappings)
func NewStatusMapper(s *zap.SugaredLogger, mappingConfig interface{}) *StatusMapper {
	mapper := &StatusMapper{
		s:       s,
		mapping: getDefaultStatusMapping(),
	}
	if mappingConfig == nil {
		return mapper
	}
	customMapping := parseMappingConfig(s, mappingConfig)
	if customMapping != nil {
		// Merge custom mapping with defaults (custom overrides defaults)
		for key, value := range customMapping {
			mapper.mapping[strings.ToLower(key)] = value
		}
		if s != nil {
			s.Infof("Loaded custom status mapping with %d entries", len(customMapping))
		}
	}

	return mapper
}

// parseMappingConfig parses the mapping configuration from various formats
func parseMappingConfig(s *zap.SugaredLogger, mappingConfig interface{}) map[string]string {
	switch v := mappingConfig.(type) {
	case string:
		// String format (from environment variable)
		return parseJSONString(s, v)
	case map[string]interface{}:
		// Map format (from JSON config file)
		return convertInterfaceMap(s, v)
	case map[string]string:
		// Direct map format
		return v
	default:
		if s != nil {
			s.Warnf("Unexpected mapping config type: %T, using defaults", mappingConfig)
		}
		return nil
	}
}

// parseJSONString parses a JSON string into a map
func parseJSONString(s *zap.SugaredLogger, jsonStr string) map[string]string {
	if len(strings.TrimSpace(jsonStr)) == 0 {
		return nil
	}
	var result map[string]string
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		if s != nil {
			s.Warnf("Failed to parse STATUS_MAPPING JSON string, using defaults: %v", err)
		}
		return nil
	}
	return result
}

// convertInterfaceMap converts map[string]interface{} to map[string]string
func convertInterfaceMap(s *zap.SugaredLogger, m map[string]interface{}) map[string]string {
	result := make(map[string]string, len(m))
	for key, value := range m {
		if strValue, ok := value.(string); ok {
			result[key] = strValue
		} else {
			if s != nil {
				s.Warnf("Skipping non-string value for key %q: %v (type: %T)", key, value, value)
			}
		}
	}
	return result
}

// MapStatus maps a database status to its classified status
// Returns the mapped status, or the original if no mapping exists
func (m *StatusMapper) MapStatus(dbStatus string) string {
	if dbStatus == "" {
		return ""
	}
	// Normalise to lowercase for lookup
	normalized := strings.ToLower(strings.TrimSpace(dbStatus))
	if mapped, exists := m.mapping[normalized]; exists {
		return mapped
	}
	// If no mapping exists, return the original value
	return dbStatus
}

// getDefaultStatusMapping returns the default status classification mapping
func getDefaultStatusMapping() map[string]string {
	return map[string]string{
		"active":      "active",
		"unlisted":    "removed",
		"yanked":      "removed",
		"deleted":     "deleted",
		"deprecated":  "deprecated",
		"unpublished": "removed",
		"archived":    "deprecated",
	}
}
