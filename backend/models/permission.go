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
	Status      string    `json:"status" gorm:"size:20;default:active"`
	Category    string    `json:"category" gorm:"size:50;default:null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	// 关联关系：权限被多个角色拥有（多对多）
	// many2many:role_permissions 指定中间表为 role_permissions
	// 可以通过 permission.Roles 直接访问拥有该权限的所有角色
	Roles []Role `json:"roles" gorm:"many2many:role_permissions;"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}
