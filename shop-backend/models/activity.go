package models

import (
	"time"
)

// Activity 活动模型
type Activity struct {
	ID          int               `json:"id" gorm:"primaryKey"`
	Name        string            `json:"name" gorm:"size:100;not null"`
	Status      string            `json:"status" gorm:"size:20;not null;default:active"`
	StartTime   time.Time         `json:"start_time" gorm:"not null"`
	EndTime     time.Time         `json:"end_time" gorm:"not null"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Products    []ActivityProduct `json:"products" gorm:"foreignKey:ActivityID"`
}

// ActivityProduct 活动商品模型
type ActivityProduct struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	ActivityID int       `json:"activity_id" gorm:"not null;index"`
	ProductID  int       `json:"product_id" gorm:"not null;index"`
	Price      float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	Stock      int       `json:"stock" gorm:"not null;default:0"`
	Limit      int       `json:"limit" gorm:"not null;default:0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName 设置表名
func (Activity) TableName() string {
	return "activities"
}

// TableName 设置表名
func (ActivityProduct) TableName() string {
	return "activity_products"
}
