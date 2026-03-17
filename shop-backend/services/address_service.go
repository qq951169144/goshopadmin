package services

import (
	"errors"
	"gorm.io/gorm"
	"shop-backend/models"
)

// AddressService 地址服务
type AddressService struct {
	db *gorm.DB
}

// NewAddressService 创建地址服务实例
func NewAddressService(db *gorm.DB) *AddressService {
	return &AddressService{db: db}
}

// AddressResponse 地址响应结构
type AddressResponse struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	DetailAddress string `json:"detail_address"`
	IsDefault     bool   `json:"is_default"`
}

// CreateAddressRequest 创建地址请求
type CreateAddressRequest struct {
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	DetailAddress string `json:"detail_address"`
	IsDefault     bool   `json:"is_default"`
}

// UpdateAddressRequest 更新地址请求
type UpdateAddressRequest struct {
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	DetailAddress string `json:"detail_address"`
	IsDefault     bool   `json:"is_default"`
}

// GetAddressList 获取地址列表
func (s *AddressService) GetAddressList(customerID int) ([]AddressResponse, error) {
	var addresses []models.Address
	result := s.db.Where("customer_id = ?", customerID).Order("is_default DESC, created_at DESC").Find(&addresses)
	if result.Error != nil {
		return nil, result.Error
	}

	var list []AddressResponse
	for _, addr := range addresses {
		list = append(list, AddressResponse{
			ID:            addr.ID,
			Name:          addr.Name,
			Phone:         addr.Phone,
			Province:      addr.Province,
			City:          addr.City,
			District:      addr.District,
			DetailAddress: addr.DetailAddress,
			IsDefault:     addr.IsDefault,
		})
	}

	return list, nil
}

// GetAddressByID 根据ID获取地址
func (s *AddressService) GetAddressByID(customerID, addressID int) (*AddressResponse, error) {
	var address models.Address
	result := s.db.Where("id = ? AND customer_id = ?", addressID, customerID).First(&address)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &AddressResponse{
		ID:            address.ID,
		Name:          address.Name,
		Phone:         address.Phone,
		Province:      address.Province,
		City:          address.City,
		District:      address.District,
		DetailAddress: address.DetailAddress,
		IsDefault:     address.IsDefault,
	}, nil
}

// CreateAddress 创建地址
func (s *AddressService) CreateAddress(customerID int, req CreateAddressRequest) (*AddressResponse, error) {
	// 如果设置为默认地址，先将其他地址设为非默认
	if req.IsDefault {
		s.db.Model(&models.Address{}).Where("customer_id = ?", customerID).Update("is_default", false)
	}

	address := models.Address{
		CustomerID:    customerID,
		Name:          req.Name,
		Phone:         req.Phone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		DetailAddress: req.DetailAddress,
		IsDefault:     req.IsDefault,
	}

	if err := s.db.Create(&address).Error; err != nil {
		return nil, err
	}

	return &AddressResponse{
		ID:            address.ID,
		Name:          address.Name,
		Phone:         address.Phone,
		Province:      address.Province,
		City:          address.City,
		District:      address.District,
		DetailAddress: address.DetailAddress,
		IsDefault:     address.IsDefault,
	}, nil
}

// UpdateAddress 更新地址
func (s *AddressService) UpdateAddress(customerID, addressID int, req UpdateAddressRequest) (*AddressResponse, error) {
	var address models.Address
	result := s.db.Where("id = ? AND customer_id = ?", addressID, customerID).First(&address)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("地址不存在")
		}
		return nil, result.Error
	}

	// 如果设置为默认地址，先将其他地址设为非默认
	if req.IsDefault && !address.IsDefault {
		s.db.Model(&models.Address{}).Where("customer_id = ?", customerID).Update("is_default", false)
	}

	// 更新字段
	if req.Name != "" {
		address.Name = req.Name
	}
	if req.Phone != "" {
		address.Phone = req.Phone
	}
	if req.Province != "" {
		address.Province = req.Province
	}
	if req.City != "" {
		address.City = req.City
	}
	if req.District != "" {
		address.District = req.District
	}
	if req.DetailAddress != "" {
		address.DetailAddress = req.DetailAddress
	}
	address.IsDefault = req.IsDefault

	if err := s.db.Save(&address).Error; err != nil {
		return nil, err
	}

	return &AddressResponse{
		ID:            address.ID,
		Name:          address.Name,
		Phone:         address.Phone,
		Province:      address.Province,
		City:          address.City,
		District:      address.District,
		DetailAddress: address.DetailAddress,
		IsDefault:     address.IsDefault,
	}, nil
}

// DeleteAddress 删除地址
func (s *AddressService) DeleteAddress(customerID, addressID int) error {
	result := s.db.Where("id = ? AND customer_id = ?", addressID, customerID).Delete(&models.Address{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("地址不存在")
	}
	return nil
}

// SetDefaultAddress 设置默认地址
func (s *AddressService) SetDefaultAddress(customerID, addressID int) error {
	var address models.Address
	result := s.db.Where("id = ? AND customer_id = ?", addressID, customerID).First(&address)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("地址不存在")
		}
		return result.Error
	}

	// 先将所有地址设为非默认
	if err := s.db.Model(&models.Address{}).Where("customer_id = ?", customerID).Update("is_default", false).Error; err != nil {
		return err
	}

	// 设置当前地址为默认
	address.IsDefault = true
	return s.db.Save(&address).Error
}

// GetDefaultAddress 获取默认地址
func (s *AddressService) GetDefaultAddress(customerID int) (*AddressResponse, error) {
	var address models.Address
	result := s.db.Where("customer_id = ? AND is_default = ?", customerID, true).First(&address)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &AddressResponse{
		ID:            address.ID,
		Name:          address.Name,
		Phone:         address.Phone,
		Province:      address.Province,
		City:          address.City,
		District:      address.District,
		DetailAddress: address.DetailAddress,
		IsDefault:     address.IsDefault,
	}, nil
}
