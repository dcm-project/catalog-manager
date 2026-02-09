package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// CatalogItemInstance represents a catalog item instance in the database
type CatalogItemInstance struct {
	ID                     string                  `gorm:"column:id;primaryKey"`
	ApiVersion             string                  `gorm:"column:api_version;not null"`
	DisplayName            string                  `gorm:"column:display_name;not null"`
	Spec                   CatalogItemInstanceSpec `gorm:"column:spec;type:jsonb;not null"`
	ServiceTypeInstanceUid string                  `gorm:"column:service_type_instance_uid"`
	Path                   string                  `gorm:"column:path;not null"`
	CreateTime             time.Time               `gorm:"column:create_time;autoCreateTime"`
	UpdateTime             time.Time               `gorm:"column:update_time;autoUpdateTime"`

	// Indexed field for filtering
	SpecCatalogItemId string `gorm:"column:spec_catalog_item_id;not null;index"`
}

// TableName specifies the table name for CatalogItemInstance
func (CatalogItemInstance) TableName() string {
	return "catalog_item_instances"
}

// CatalogItemInstanceList is a slice of CatalogItemInstance for list results
type CatalogItemInstanceList []CatalogItemInstance

// CatalogItemInstanceSpec represents the spec field of a catalog item instance
type CatalogItemInstanceSpec struct {
	CatalogItemId string      `json:"catalog_item_id"`
	UserValues    []UserValue `json:"user_values"`
}

// Scan implements sql.Scanner for CatalogItemInstanceSpec
func (s *CatalogItemInstanceSpec) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("catalog item instance spec cannot be null")
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer for CatalogItemInstanceSpec
func (s CatalogItemInstanceSpec) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// UserValue represents a user-provided value for a field
type UserValue struct {
	Path  string `json:"path"`
	Value any    `json:"value"`
}
