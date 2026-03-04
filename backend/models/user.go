package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"size:50;uniqueIndex;not null"`
	Password  string    `json:"password" gorm:"size:100;not null"`
	RoleID    int       `json:"role_id" gorm:"not null"`
	Status    string    `json:"status" gorm:"default:active"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	// 关联关系：用户属于一个角色（多对一）
	// foreignKey:RoleID 指定外键字段为 RoleID
	// 可以通过 user.Role 直接访问用户所属的角色信息
	Role Role `json:"role" gorm:"foreignKey:RoleID"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
