package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ServiceType represents a service type definition in the database
type ServiceType struct {
	ID          string    `gorm:"column:id;primaryKey"`
	ApiVersion  string    `gorm:"column:api_version;not null"`
	ServiceType string    `gorm:"column:service_type;not null;uniqueIndex"`
	Metadata    Metadata  `gorm:"column:metadata;type:jsonb"`
	Spec        JSONMap   `gorm:"column:spec;type:jsonb;not null"`
	Path        string    `gorm:"column:path;not null"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime"`
}

type ServiceTypeList []ServiceType

// TableName specifies the table name for ServiceType
func (ServiceType) TableName() string {
	return "service_types"
}

// Metadata represents the metadata field with labels
type Metadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}

// Scan implements sql.Scanner for Metadata
func (m *Metadata) Scan(value any) error {
	if value == nil {
		m.Labels = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer for Metadata
func (m Metadata) Value() (driver.Value, error) {
	if m.Labels == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// JSONMap represents an arbitrary JSON object
type JSONMap map[string]any

// Scan implements sql.Scanner for JSONMap
func (j *JSONMap) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	result := make(map[string]any)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}

// Value implements driver.Valuer for JSONMap
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}
