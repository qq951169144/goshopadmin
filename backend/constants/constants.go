package constants

// Status 状态常量
const (
	// StatusActive 激活状态
	StatusActive = "active"
	
	// StatusInactive 禁用状态
	StatusInactive = "inactive"
)

// AuditStatus 审核状态常量
const (
	// AuditStatusPending 待审核状态
	AuditStatusPending = "pending"
	
	// AuditStatusApproved 已通过状态
	AuditStatusApproved = "approved"
	
	// AuditStatusRejected 已拒绝状态
	AuditStatusRejected = "rejected"
)

// MerchantRole 商户角色常量
const (
	// MerchantRoleOwner 老板角色
	MerchantRoleOwner = "owner"
	
	// MerchantRoleManager 经理角色
	MerchantRoleManager = "manager"
	
	// MerchantRoleStaff 员工角色
	MerchantRoleStaff = "staff"
)

// OrderStatus 订单状态常量
const (
	// OrderStatusPending 待支付状态
	OrderStatusPending = "pending"
	
	// OrderStatusProcessing 处理中状态
	OrderStatusProcessing = "processing"
	
	// OrderStatusCompleted 已完成状态
	OrderStatusCompleted = "completed"
	
	// OrderStatusFailed 失败状态
	OrderStatusFailed = "failed"
)

// PaymentStatus 支付状态常量
const (
	// PaymentStatusPending 待支付状态
	PaymentStatusPending = "pending"
	
	// PaymentStatusSuccess 支付成功状态
	PaymentStatusSuccess = "success"
	
	// PaymentStatusFailed 支付失败状态
	PaymentStatusFailed = "failed"
)

// ActivityStatus 活动状态常量
const (
	// ActivityStatusActive 活动激活状态
	ActivityStatusActive = "active"
	
	// ActivityStatusInactive 活动禁用状态
	ActivityStatusInactive = "inactive"
)

// MerchantStatementStatus 商户对账单状态常量
const (
	// MerchantStatementStatusDraft 草稿状态
	MerchantStatementStatusDraft = "draft"
	
	// MerchantStatementStatusConfirmed 已确认状态
	MerchantStatementStatusConfirmed = "confirmed"
	
	// MerchantStatementStatusSettled 已结算状态
	MerchantStatementStatusSettled = "settled"
)

// WithdrawStatus 提现状态常量
const (
	// WithdrawStatusPending 待处理状态
	WithdrawStatusPending = "pending"
	
	// WithdrawStatusProcessing 处理中状态
	WithdrawStatusProcessing = "processing"
	
	// WithdrawStatusCompleted 已完成状态
	WithdrawStatusCompleted = "completed"
	
	// WithdrawStatusFailed 失败状态
	WithdrawStatusFailed = "failed"
)

// ShippingStatus 物流状态常量
const (
	// ShippingStatusPending 待发货状态
	ShippingStatusPending = "pending"
	
	// ShippingStatusShipped 已发货状态
	ShippingStatusShipped = "shipped"
	
	// ShippingStatusDelivered 已送达状态
	ShippingStatusDelivered = "delivered"
	
	// ShippingStatusReturned 已退回状态
	ShippingStatusReturned = "returned"
)

// ActivityType 活动类型常量
const (
	// ActivityTypeSeckill 秒杀活动
	ActivityTypeSeckill = "seckill"
	
	// ActivityTypeRedeemCode 兑换码活动
	ActivityTypeRedeemCode = "redeem_code"
)

// RedeemCodeStatus 兑换码状态常量
const (
	// RedeemCodeStatusUnused 未使用
	RedeemCodeStatusUnused = "unused"
	
	// RedeemCodeStatusUsed 已使用
	RedeemCodeStatusUsed = "used"
	
	// RedeemCodeStatusExpired 已过期
	RedeemCodeStatusExpired = "expired"
	
	// RedeemCodeStatusDisabled 已禁用
	RedeemCodeStatusDisabled = "disabled"
)

// RedeemCodeType 兑换码类型常量
const (
	// RedeemCodeTypeNumeric 纯数字
	RedeemCodeTypeNumeric = "numeric"
	
	// RedeemCodeTypeAlpha 纯字母
	RedeemCodeTypeAlpha = "alpha"
	
	// RedeemCodeTypeAlphanumeric 数字+字母
	RedeemCodeTypeAlphanumeric = "alphanumeric"
)

// RedeemCodeLogStatus 兑换码核销状态常量
const (
	// RedeemCodeLogStatusPending 待核销
	RedeemCodeLogStatusPending = "pending"
	
	// RedeemCodeLogStatusVerified 已核销
	RedeemCodeLogStatusVerified = "verified"
	
	// RedeemCodeLogStatusRejected 已拒绝
	RedeemCodeLogStatusRejected = "rejected"
)

// SKUActivityType SKU活动类型常量
const (
	// SKUActivityTypeNormal 普通SKU
	SKUActivityTypeNormal = 0
	
	// SKUActivityTypeActivity 活动专用SKU
	SKUActivityTypeActivity = 1
)