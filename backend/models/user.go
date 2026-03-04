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
	Role      Role      `json:"role" gorm:"foreignKey:RoleID"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
