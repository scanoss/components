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
	"encoding/json"
	"strings"

	"go.uber.org/zap"
)

// StatusMapper handles mapping of database statuses to classified statuses
type StatusMapper struct {
	mapping map[string]string
	s       *zap.SugaredLogger
}

// NewStatusMapper creates a new StatusMapper with the provided mapping JSON string
// If mappingJSON is empty or invalid, it uses default mappings
func NewStatusMapper(s *zap.SugaredLogger, mappingJSON string) *StatusMapper {
	mapper := &StatusMapper{
		s:       s,
		mapping: getDefaultStatusMapping(),
	}

	// If custom mapping provided, try to parse it
	if len(strings.TrimSpace(mappingJSON)) > 0 {
		var customMapping map[string]string
		err := json.Unmarshal([]byte(mappingJSON), &customMapping)
		if err != nil {
			s.Warnf("Failed to parse STATUS_MAPPING JSON, using defaults: %v", err)
		} else {
			// Merge custom mapping with defaults (custom overrides defaults)
			for key, value := range customMapping {
				mapper.mapping[strings.ToLower(key)] = value
			}
			s.Infof("Loaded custom status mapping with %d entries", len(customMapping))
		}
	}

	return mapper
}

// MapStatus maps a database status to its classified status
// Returns the mapped status, or the original if no mapping exists
func (m *StatusMapper) MapStatus(dbStatus string) string {
	if dbStatus == "" {
		return ""
	}

	// Normalize to lowercase for lookup
	normalized := strings.ToLower(strings.TrimSpace(dbStatus))

	if mapped, exists := m.mapping[normalized]; exists {
		return mapped
	}

	// If no mapping exists, return original value
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
