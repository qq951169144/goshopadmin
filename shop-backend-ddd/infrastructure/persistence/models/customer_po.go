package models

// CustomerPO 客户持久化对象
type CustomerPO struct {
	ID       int    `gorm:"primaryKey;autoIncrement"`
	Username string `gorm:"column:username;size:50;unique;not null"`
	Password string `gorm:"column:password;size:100;not null"`
	Status   string `gorm:"column:status;type:enum('active','inactive');default:'active'"`
}

func (CustomerPO) TableName() string { return "customers" }

// AddressPO 地址持久化对象
type AddressPO struct {
	ID            int    `gorm:"primaryKey;autoIncrement"`
	CustomerID    int    `gorm:"column:customer_id;not null;index"`
	Name          string `gorm:"column:name;size:50;not null"`
	Phone         string `gorm:"column:phone;size:20;not null"`
	Province      string `gorm:"column:province;size:50"`
	City          string `gorm:"column:city;size:50"`
	District      string `gorm:"column:district;size:50"`
	DetailAddress string `gorm:"column:detail_address;size:200;not null"`
}

func (AddressPO) TableName() string { return "addresses" }
