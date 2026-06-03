package customer

// Customer 客户实体
type Customer struct {
	id       int
	username string
	status   string
}

// NewCustomer 创建客户
func NewCustomer(id int, username string) *Customer {
	return &Customer{id: id, username: username, status: "active"}
}

func (c *Customer) ID() int         { return c.id }
func (c *Customer) Username() string { return c.username }
func (c *Customer) Status() string   { return c.status }

// Address 地址值对象
type Address struct {
	id            int
	customerID    int
	name          string
	phone         string
	province      string
	city          string
	district      string
	detailAddress string
}

// NewAddress 创建地址
func NewAddress(id, customerID int, name, phone, province, city, district, detail string) Address {
	return Address{
		id: id, customerID: customerID, name: name, phone: phone,
		province: province, city: city, district: district, detailAddress: detail,
	}
}

func (a Address) ID() int             { return a.id }
func (a Address) CustomerID() int     { return a.customerID }
func (a Address) Name() string        { return a.name }
func (a Address) Phone() string       { return a.phone }
func (a Address) Province() string    { return a.province }
func (a Address) City() string        { return a.city }
func (a Address) District() string    { return a.district }
func (a Address) DetailAddress() string { return a.detailAddress }
