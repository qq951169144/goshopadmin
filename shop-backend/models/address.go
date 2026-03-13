package models

import (
	"time"
)

// Address 收货地址模型
type Address struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	CustomerID    uint      `json:"customer_id" gorm:"not null;index"`
	Name          string    `json:"name" gorm:"size:50;not null"`
	Phone         string    `json:"phone" gorm:"size:20;not null"`
	Province      string    `json:"province" gorm:"size:50;not null"`
	City          string    `json:"city" gorm:"size:50;not null"`
	District      string    `json:"district" gorm:"size:50;not null"`
	DetailAddress string    `json:"detail_address" gorm:"size:255;not null"`
	IsDefault     bool      `json:"is_default" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 设置表名
func (Address) TableName() string {
	return "addresses"
}
