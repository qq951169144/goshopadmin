package models

import (
	"time"
)

// RedeemCode 兑换码模型
type RedeemCode struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	Code       string    `json:"code" gorm:"size:50;not null;uniqueIndex"`
	ActivityID int       `json:"activity_id" gorm:"not null;index"`
	Value      float64   `json:"value" gorm:"type:decimal(10,2);not null"`
	Status     string    `json:"status" gorm:"size:20;not null;default:active"`
	ExpireTime time.Time `json:"expire_time" gorm:"not null"`
	UsedAt     *time.Time `json:"used_at"`
	CustomerID int       `json:"customer_id" gorm:"default:0;index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// RedeemCodeLog 兑换码使用记录模型
type RedeemCodeLog struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	ActivityID  int       `json:"activity_id" gorm:"not null;index"`
	RedeemCodeID int      `json:"redeem_code_id" gorm:"not null;index"`
	CustomerID  int       `json:"customer_id" gorm:"not null;index"`
	Code        string    `json:"code" gorm:"size:50;not null"`
	Value       float64   `json:"value" gorm:"type:decimal(10,2);not null"`
	RedeemTime  time.Time `json:"redeem_time" gorm:"not null"`
	Status      string    `json:"status" gorm:"size:20;not null;default:success"`
	CreatedAt   time.Time `json:"created_at"`
}

// ActivityRedeemSetting 活动兑换设置模型
type ActivityRedeemSetting struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	ActivityID  int       `json:"activity_id" gorm:"not null;uniqueIndex"`
	Value       float64   `json:"value" gorm:"type:decimal(10,2);not null"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 设置表名
func (RedeemCode) TableName() string {
	return "redeem_codes"
}

// TableName 设置表名
func (RedeemCodeLog) TableName() string {
	return "redeem_code_logs"
}

// TableName 设置表名
func (ActivityRedeemSetting) TableName() string {
	return "activity_redeem_settings"
}
