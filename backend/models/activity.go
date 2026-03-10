package models

import (
	"time"
)

// Activity 活动模型
type Activity struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	MerchantID int       `json:"merchant_id" gorm:"not null"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Type      string    `json:"type" gorm:"size:20;not null"`
	StartTime time.Time `json:"start_time" gorm:"not null"`
	EndTime   time.Time `json:"end_time" gorm:"not null"`
	Status    string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedBy int       `json:"created_by" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Rules    []ActivityRule    `json:"rules" gorm:"foreignKey:ActivityID"`
	Products []ActivityProduct `json:"products" gorm:"foreignKey:ActivityID"`
}

// ActivityRule 活动规则模型
type ActivityRule struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	ActivityID int       `json:"activity_id" gorm:"not null"`
	RuleType   string    `json:"rule_type" gorm:"size:20;not null"`
	RuleValue  string    `json:"rule_value" gorm:"type:json;not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联
	Activity Activity `json:"activity" gorm:"foreignKey:ActivityID"`
}

// ActivityProduct 活动商品模型
type ActivityProduct struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	ActivityID   int       `json:"activity_id" gorm:"not null"`
	ProductID    int       `json:"product_id" gorm:"not null"`
	SKUID        int       `json:"sku_id" gorm:"not null"`
	MerchantID   int       `json:"merchant_id" gorm:"not null"`
	OriginalPrice float64   `json:"original_price" gorm:"type:decimal(10,2);not null"`
	ActivityPrice float64   `json:"activity_price" gorm:"type:decimal(10,2);not null"`
	Stock        int       `json:"stock" gorm:"not null"`
	Status       string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联
	Activity Activity   `json:"activity" gorm:"foreignKey:ActivityID"`
	Product  Product    `json:"product" gorm:"foreignKey:ProductID"`
	SKU      ProductSKU `json:"sku" gorm:"foreignKey:SKUID"`
}

// ActivityStat 活动效果统计模型
type ActivityStat struct {
	ID             int       `json:"id" gorm:"primaryKey"`
	ActivityID     int       `json:"activity_id" gorm:"not null"`
	ViewCount      int       `json:"view_count" gorm:"default:0"`
	ParticipantCount int     `json:"participant_count" gorm:"default:0"`
	OrderCount     int       `json:"order_count" gorm:"default:0"`
	TotalAmount    float64   `json:"total_amount" gorm:"type:decimal(10,2);default:0"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 关联
	// Activity Activity `json:"activity" gorm:"foreignKey:ActivityID"`
}
