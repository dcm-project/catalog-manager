package model

import (
	"time"
)

// ServiceType represents a service type definition in the database
type ServiceType struct {
	ID          string    `gorm:"column:id;primaryKey"`
	ApiVersion  string    `gorm:"column:api_version;not null"`
	ServiceType string    `gorm:"column:service_type;not null;uniqueIndex"`
	Metadata    Metadata  `gorm:"column:metadata;type:jsonb;serializer:json"`
	Spec        JSONMap   `gorm:"column:spec;type:jsonb;not null;serializer:json"`
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

// JSONMap represents an arbitrary JSON object
type JSONMap map[string]any
