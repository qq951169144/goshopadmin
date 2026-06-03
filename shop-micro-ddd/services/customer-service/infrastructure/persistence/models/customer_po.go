package models

import "time"

// CustomerPO 客户持久化对象
type CustomerPO struct {
	ID        int          `gorm:"primaryKey;type:int;autoIncrement"`
	Username  string       `gorm:"column:username;type:varchar(50);unique;not null"`
	CreatedAt time.Time    `gorm:"column:created_at;autoCreateTime"`
	Addresses []AddressPO  `gorm:"foreignKey:CustomerID"`
}

// TableName 客户表名
func (CustomerPO) TableName() string { return "customers" }

// AddressPO 地址持久化对象
type AddressPO struct {
	ID         int    `gorm:"primaryKey;type:int;autoIncrement"`
	CustomerID int    `gorm:"column:customer_id;type:int;not null;index"`
	MerchantID int    `gorm:"column:merchant_id;type:int;not null"`
	Detail     string `gorm:"column:detail;type:varchar(500);not null"`
}

// TableName 地址表名
func (AddressPO) TableName() string { return "customer_addresses" }
