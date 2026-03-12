package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"size:191;uniqueIndex;not null"`
	Password  string    `json:"password" gorm:"type:longtext;not null"`
	RoleID    int       `json:"role_id" gorm:"not null"`
	Status    string    `json:"status" gorm:"default:active"`
	CreatedAt time.Time `json:"created_at" gorm:"type:datetime(3)"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:datetime(3)"`
	Email     string    `json:"email" gorm:"size:191;uniqueIndex"`
	// 关联关系：用户属于一个角色（多对一）
	// foreignKey:RoleID 指定外键字段为 RoleID
	// 可以通过 user.Role 直接访问用户所属的角色信息
	Role Role `json:"role" gorm:"foreignKey:RoleID"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
