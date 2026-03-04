package models

import (
	"time"
)

// Role 角色模型
type Role struct {
	ID          int           `json:"id" gorm:"primaryKey"`
	Name        string        `json:"name" gorm:"size:50;not null"`
	Description string        `json:"description" gorm:"size:200"`
	CreatedAt   time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	Permissions []Permission  `json:"permissions" gorm:"many2many:role_permissions;"`
	Users       []User        `json:"users" gorm:"foreignKey:RoleID"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}
