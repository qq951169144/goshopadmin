package models

import (
	"time"
)

// Activity 活动模型
type Activity struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	MerchantID int       `json:"merchant_id" gorm:"not null"`
	Name       string    `json:"name" gorm:"size:100;not null"`
	Type       string    `json:"type" gorm:"type:enum('seckill','redeem_code');not null"`
	StartTime  time.Time `json:"start_time" gorm:"not null"`
	EndTime    time.Time `json:"end_time" gorm:"not null"`
	Status     string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedBy  int       `json:"created_by" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联
	Products      []ActivityProduct      `json:"products" gorm:"foreignKey:ActivityID"`
	RedeemSetting *ActivityRedeemSetting `json:"redeem_setting" gorm:"foreignKey:ActivityID"`
	RedeemCodes   []RedeemCode           `json:"redeem_codes" gorm:"foreignKey:ActivityID"`
}

// ActivityProduct 活动商品模型
type ActivityProduct struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	ActivityID    int       `json:"activity_id" gorm:"not null"`
	ProductID     int       `json:"product_id" gorm:"not null"`
	SkuID         int       `json:"sku_id" gorm:"column:sku_id;not null"`
	MerchantID    int       `json:"merchant_id" gorm:"not null"`
	OriginalPrice float64   `json:"original_price" gorm:"type:decimal(10,2);not null"`
	ActivityPrice float64   `json:"activity_price" gorm:"type:decimal(10,2);not null"`
	Stock         int       `json:"stock" gorm:"not null"`
	ProductType   string    `json:"product_type" gorm:"type:enum('seckill','redeem');default:'seckill'"`
	Status        string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联
	Activity Activity   `json:"activity" gorm:"foreignKey:ActivityID"`
	Product  Product    `json:"product" gorm:"foreignKey:ProductID"`
	Sku      ProductSku `json:"sku" gorm:"foreignKey:SkuID"`
}

// ActivityRedeemSetting 兑换码活动配置模型
type ActivityRedeemSetting struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	ActivityID    int       `json:"activity_id" gorm:"not null;uniqueIndex"`
	MerchantID    int       `json:"merchant_id" gorm:"not null"`
	CodeType      string    `json:"code_type" gorm:"size:20;not null;default:'alphanumeric'"`
	CodeLength    int       `json:"code_length" gorm:"not null;default:8"`
	ExcludeChars  string    `json:"exclude_chars" gorm:"size:20;default:'IO10'"`
	TotalQuantity int       `json:"total_quantity" gorm:"not null;default:0"`
	LimitPerUser  int       `json:"limit_per_user" gorm:"default:1"`
	NeedVerify    int       `json:"need_verify" gorm:"type:tinyint(1);default:1"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联
	Activity Activity `json:"activity" gorm:"foreignKey:ActivityID"`
}

// TableName 设置表名
func (ActivityRedeemSetting) TableName() string {
	return "activity_redeem_settings"
}

// RedeemCode 兑换码模型
type RedeemCode struct {
	ID             int       `json:"id" gorm:"primaryKey"`
	ActivityID     int       `json:"activity_id" gorm:"not null"`
	MerchantID     int       `json:"merchant_id" gorm:"not null"`
	Code           string    `json:"code" gorm:"size:50;not null;uniqueIndex"`
	Status         string    `json:"status" gorm:"type:enum('unused','used','expired','disabled');default:'unused'"`
	UsedBy         int       `json:"used_by" gorm:"default:null"`
	UsedAt         time.Time `json:"used_at" gorm:"default:null"`
	ValidStartTime time.Time `json:"valid_start_time" gorm:"default:null"`
	ValidEndTime   time.Time `json:"valid_end_time" gorm:"default:null"`
	CreatedBy      int       `json:"created_by" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 关联
	Activity Activity `json:"activity" gorm:"foreignKey:ActivityID"`
}

// TableName 设置表名
func (RedeemCode) TableName() string {
	return "redeem_codes"
}

// RedeemCodeLog 兑换码核销记录模型
type RedeemCodeLog struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	RedeemCodeID int       `json:"redeem_code_id" gorm:"not null"`
	ActivityID   int       `json:"activity_id" gorm:"not null"`
	MerchantID   int       `json:"merchant_id" gorm:"not null"`
	CustomerID   int       `json:"customer_id" gorm:"not null"`
	Code         string    `json:"code" gorm:"size:50;not null"`
	VerifyBy     int       `json:"verify_by" gorm:"default:null"`
	VerifyAt     time.Time `json:"verify_at" gorm:"default:null"`
	Status       string    `json:"status" gorm:"type:enum('pending','verified','rejected');default:'pending'"`
	Remark       string    `json:"remark" gorm:"size:255;default:null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联
	RedeemCode RedeemCode `json:"redeem_code" gorm:"foreignKey:RedeemCodeID"`
	Activity   Activity   `json:"activity" gorm:"foreignKey:ActivityID"`
}

// TableName 设置表名
func (RedeemCodeLog) TableName() string {
	return "redeem_code_logs"
}

// ActivityStat 活动效果统计模型
type ActivityStat struct {
	ID               int       `json:"id" gorm:"primaryKey"`
	ActivityID       int       `json:"activity_id" gorm:"not null"`
	ViewCount        int       `json:"view_count" gorm:"default:0"`
	ParticipantCount int       `json:"participant_count" gorm:"default:0"`
	OrderCount       int       `json:"order_count" gorm:"default:0"`
	TotalAmount      float64   `json:"total_amount" gorm:"type:decimal(10,2);default:0"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// 关联
	// Activity Activity `json:"activity" gorm:"foreignKey:ActivityID"`
}
