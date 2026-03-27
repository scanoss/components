package dtos

import (
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
)

// ComponentStatusOutput represents the status information for a single component.
type ComponentStatusOutput struct {
	Purl            string               `json:"purl"`
	Name            string               `json:"name"`
	Requirement     string               `json:"requirement,omitempty"`
	VersionStatus   *VersionStatusOutput `json:"version_status,omitempty"`
	ComponentStatus *ComponentStatusInfo `json:"component_status,omitempty"`
}

// VersionStatusOutput represents the status of a specific version.
type VersionStatusOutput struct {
	Version          string             `json:"version"`
	Status           string             `json:"status"`
	RepositoryStatus string             `json:"repository_status,omitempty"`
	IndexedDate      string             `json:"indexed_date,omitempty"`
	StatusChangeDate string             `json:"status_change_date,omitempty"`
	ErrorMessage     *string            `json:"error_message,omitempty"`
	ErrorCode        *domain.StatusCode `json:"error_code,omitempty"`
}

// ComponentStatusInfo represents the status of a component (ignoring version).
type ComponentStatusInfo struct {
	Status           string             `json:"status"`
	RepositoryStatus string             `json:"repository_status,omitempty"`
	FirstIndexedDate string             `json:"first_indexed_date,omitempty"`
	LastIndexedDate  string             `json:"last_indexed_date,omitempty"`
	StatusChangeDate string             `json:"status_change_date,omitempty"`
	ErrorMessage     *string            `json:"error_message,omitempty"`
	ErrorCode        *domain.StatusCode `json:"error_code,omitempty"`
}

// ComponentsStatusOutput represents the status information for multiple components.
type ComponentsStatusOutput struct {
	Components []ComponentStatusOutput `json:"components"`
}

// StringPtr returns a pointer to the provided string value
// This is useful for optional string fields that require pointers.
func StringPtr(s string) *string {
	return &s
}
