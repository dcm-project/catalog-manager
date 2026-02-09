package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// CatalogItem represents a catalog item in the database
type CatalogItem struct {
	ID          string          `gorm:"column:id;primaryKey"`
	ApiVersion  string          `gorm:"column:api_version;not null"`
	DisplayName string          `gorm:"column:display_name;not null"`
	Spec        CatalogItemSpec `gorm:"column:spec;type:jsonb;not null"`
	Path        string          `gorm:"column:path;not null"`
	CreateTime  time.Time       `gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time       `gorm:"column:update_time;autoUpdateTime"`

	// Indexed field for filtering
	SpecServiceType string `gorm:"column:spec_service_type;not null;index"`
}

// TableName specifies the table name for CatalogItem
func (CatalogItem) TableName() string {
	return "catalog_items"
}

// CatalogItemList is a slice of CatalogItem for list results
type CatalogItemList []CatalogItem

// CatalogItemSpec represents the spec field of a catalog item
type CatalogItemSpec struct {
	ServiceType string               `json:"service_type"`
	Fields      []FieldConfiguration `json:"fields"`
}

// Scan implements sql.Scanner for CatalogItemSpec
func (s *CatalogItemSpec) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("catalog item spec cannot be null")
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer for CatalogItemSpec
func (s CatalogItemSpec) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// FieldConfiguration represents a field configuration within a catalog item
type FieldConfiguration struct {
	Path             string         `json:"path"`
	DisplayName      string         `json:"display_name,omitempty"`
	Editable         bool           `json:"editable"`
	Default          any            `json:"default,omitempty"`
	ValidationSchema map[string]any `json:"validation_schema,omitempty"`
}
