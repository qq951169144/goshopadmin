package models

import (
	"time"
)

// Role 角色模型
type Role struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:50;not null"`
	Description string    `json:"description" gorm:"size:200"`
	Status      string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	// 关联关系：角色拥有多个权限（多对多）
	// many2many:role_permissions 指定中间表为 role_permissions
	// 可以通过 role.Permissions 直接访问角色拥有的所有权限
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	// 关联关系：角色拥有多个用户（一对多）
	// foreignKey:RoleID 指定外键字段为 RoleID
	// 可以通过 role.Users 直接访问拥有该角色的所有用户
	Users []User `json:"users" gorm:"foreignKey:RoleID"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}
