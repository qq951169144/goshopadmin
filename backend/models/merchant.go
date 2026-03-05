package models

import (
	"time"
)

// Merchant 商户模型
type Merchant struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	Name            string    `json:"name" gorm:"size:100;not null"`
	ContactPerson   string    `json:"contact_person" gorm:"size:50;not null"`
	ContactPhone    string    `json:"contact_phone" gorm:"size:20;not null"`
	Address         string    `json:"address" gorm:"size:255;not null"`
	AuditStatus     string    `json:"audit_status" gorm:"default:pending"`
	Status          string    `json:"status" gorm:"default:inactive"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	ApprovedAt      *time.Time `json:"approved_at"`
	ApprovedBy      *int       `json:"approved_by"`
	// 关联关系
	Users    []MerchantUser    `json:"users" gorm:"foreignKey:MerchantID"`
	Audits   []MerchantAudit   `json:"audits" gorm:"foreignKey:MerchantID"`
	Banks    []MerchantBank    `json:"banks" gorm:"foreignKey:MerchantID"`
	Withdraw []MerchantWithdraw `json:"withdraw" gorm:"foreignKey:MerchantID"`
	Statements []MerchantStatement `json:"statements" gorm:"foreignKey:MerchantID"`
}

// TableName 指定表名
func (Merchant) TableName() string {
	return "merchants"
}
