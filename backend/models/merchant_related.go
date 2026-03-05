package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSON 类型，用于处理 JSON 字段
type JSON map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// MerchantUser 商户用户关联模型
type MerchantUser struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	MerchantID int       `json:"merchant_id" gorm:"not null"`
	UserID     int       `json:"user_id" gorm:"not null"`
	Role       string    `json:"role" gorm:"size:20;not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	// 关联关系
	Merchant Merchant `json:"merchant" gorm:"foreignKey:MerchantID"`
	User     User     `json:"user" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (MerchantUser) TableName() string {
	return "merchant_users"
}

// MerchantAudit 商户审核模型
type MerchantAudit struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	MerchantID int       `json:"merchant_id" gorm:"not null"`
	AuditType string    `json:"audit_type" gorm:"size:20;not null"`
	OldData   *JSON     `json:"old_data"`
	NewData   *JSON     `json:"new_data"`
	Status    string    `json:"status" gorm:"default:pending"`
	Remark    string    `json:"remark" gorm:"type:text"`
	CreatedBy int       `json:"created_by" gorm:"not null"`
	AuditedBy *int      `json:"audited_by"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	AuditedAt *time.Time `json:"audited_at"`
	// 关联关系
	Merchant Merchant `json:"merchant" gorm:"foreignKey:MerchantID"`
}

// TableName 指定表名
func (MerchantAudit) TableName() string {
	return "merchant_audit"
}

// MerchantBank 商户银行信息模型
type MerchantBank struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	MerchantID   int       `json:"merchant_id" gorm:"not null"`
	BankName     string    `json:"bank_name" gorm:"size:100;not null"`
	AccountName  string    `json:"account_name" gorm:"size:100;not null"`
	AccountNumber string    `json:"account_number" gorm:"size:50;not null"`
	Branch       string    `json:"branch" gorm:"size:100;not null"`
	IsDefault    bool      `json:"is_default" gorm:"default:false"`
	Status       string    `json:"status" gorm:"default:active"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	// 关联关系
	Merchant Merchant `json:"merchant" gorm:"foreignKey:MerchantID"`
}

// TableName 指定表名
func (MerchantBank) TableName() string {
	return "merchant_bank"
}

// MerchantWithdraw 商户提现记录模型
type MerchantWithdraw struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	MerchantID  int       `json:"merchant_id" gorm:"not null"`
	Amount      float64   `json:"amount" gorm:"type:decimal(10,2);not null"`
	BankID      int       `json:"bank_id" gorm:"not null"`
	Status      string    `json:"status" gorm:"default:pending"`
	OrderNo     string    `json:"order_no" gorm:"size:32;not null"`
	Remark      string    `json:"remark" gorm:"type:text"`
	CreatedBy   int       `json:"created_by" gorm:"not null"`
	ProcessedBy *int      `json:"processed_by"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	ProcessedAt *time.Time `json:"processed_at"`
	// 关联关系
	Merchant Merchant    `json:"merchant" gorm:"foreignKey:MerchantID"`
	Bank     MerchantBank `json:"bank" gorm:"foreignKey:BankID"`
}

// TableName 指定表名
func (MerchantWithdraw) TableName() string {
	return "merchant_withdraw"
}

// MerchantStatement 商户对账单模型
type MerchantStatement struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	MerchantID   int       `json:"merchant_id" gorm:"not null"`
	StatementNo  string    `json:"statement_no" gorm:"size:32;not null"`
	StartDate    time.Time `json:"start_date" gorm:"type:date;not null"`
	EndDate      time.Time `json:"end_date" gorm:"type:date;not null"`
	TotalAmount  float64   `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	OrderCount   int       `json:"order_count" gorm:"not null"`
	Status       string    `json:"status" gorm:"default:draft"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	// 关联关系
	Merchant Merchant `json:"merchant" gorm:"foreignKey:MerchantID"`
}

// TableName 指定表名
func (MerchantStatement) TableName() string {
	return "merchant_statement"
}
