package models

import (
	"time"
)

// Customer 客户模型（对应 customers 表）
type Customer struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Username    string    `json:"username" gorm:"size:50;unique;not null"`
	Password    string    `json:"password" gorm:"size:100;not null"`
	Phone       string    `json:"phone" gorm:"size:20;unique;not null"`
	Email       string    `json:"email" gorm:"size:100;unique"`
	Status      string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	Nickname    string    `json:"nickname" gorm:"size:50"`
	Avatar      string    `json:"avatar" gorm:"size:255"`
	LastLoginAt time.Time `json:"last_login_at"`
	LastLoginIp string    `json:"last_login_ip" gorm:"size:50"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 设置表名
func (Customer) TableName() string {
	return "customers"
}
