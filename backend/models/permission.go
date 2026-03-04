package models

import (
	"time"
)

// Permission 权限模型
type Permission struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:50;not null"`
	Code        string    `json:"code" gorm:"size:50;not null"`
	Description string    `json:"description" gorm:"size:200"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Roles       []Role    `json:"roles" gorm:"many2many:role_permissions;"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}
