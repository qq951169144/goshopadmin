package customer

import "errors"

var (
	// ErrCustomerNotFound 客户不存在
	ErrCustomerNotFound = errors.New("客户不存在")
	// ErrAddressNotFound 地址不存在
	ErrAddressNotFound = errors.New("地址不存在")
	// ErrAddressNotBelongToCustomer 地址不属于该客户
	ErrAddressNotBelongToCustomer = errors.New("地址不属于该客户")
)

// Customer 客户实体
type Customer struct {
	id        int
	username  string
	addresses []Address
}

// Address 地址值对象
type Address struct {
	id         int
	merchantID int
	detail     string
}

// NewCustomer 创建客户
func NewCustomer(id int, username string, addresses []Address) *Customer {
	return &Customer{id: id, username: username, addresses: addresses}
}

// NewAddress 创建地址
func NewAddress(id, merchantID int, detail string) Address {
	return Address{id: id, merchantID: merchantID, detail: detail}
}

// VerifyAddress 验证地址是否属于该客户，返回商户ID
func (c *Customer) VerifyAddress(addressID int) (int, error) {
	for _, addr := range c.addresses {
		if addr.id == addressID {
			return addr.merchantID, nil
		}
	}
	return 0, ErrAddressNotBelongToCustomer
}

// Getter 方法
func (c *Customer) ID() int              { return c.id }
func (c *Customer) Username() string     { return c.username }
func (c *Customer) Addresses() []Address { return c.addresses }

func (a Address) ID() int         { return a.id }
func (a Address) MerchantID() int { return a.merchantID }
func (a Address) Detail() string  { return a.detail }
