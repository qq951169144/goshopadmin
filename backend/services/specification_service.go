package services

import (
	"errors"
	"goshopadmin/models"

	"gorm.io/gorm"
)

// SpecificationService 规格服务
type SpecificationService struct {
	DB *gorm.DB
}

// NewSpecificationService 创建规格服务实例
func NewSpecificationService(db *gorm.DB) *SpecificationService {
	return &SpecificationService{DB: db}
}

// CreateSpecification 创建商品规格
func (s *SpecificationService) CreateSpecification(spec *models.ProductSpecification, merchantID int) error {
	// 检查商品是否属于该商户
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", spec.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	result = s.DB.Create(spec)
	return result.Error
}

// UpdateSpecification 更新规格
func (s *SpecificationService) UpdateSpecification(specID int, merchantID int, name string, sortOrder int) error {
	// 检查规格所属的商品是否属于该商户
	var spec models.ProductSpecification
	result := s.DB.First(&spec, specID)
	if result.Error != nil {
		return errors.New("规格不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", spec.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 更新规格
	result = s.DB.Model(&spec).Updates(map[string]interface{}{
		"name": name,
		"sort": sortOrder,
	})
	return result.Error
}

// DeleteSpecification 删除规格
func (s *SpecificationService) DeleteSpecification(specID int, merchantID int) error {
	// 检查规格所属的商品是否属于该商户
	var spec models.ProductSpecification
	result := s.DB.First(&spec, specID)
	if result.Error != nil {
		return errors.New("规格不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", spec.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 检查规格是否被SKU使用
	var count int64
	s.DB.Model(&models.ProductSkuSpec{}).Where("spec_id = ?", specID).Count(&count)
	if count > 0 {
		return errors.New("规格已被SKU使用，无法删除")
	}

	// 开启事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除规格下的所有规格值
	if err := tx.Where("spec_id = ?", specID).Delete(&models.ProductSpecificationValue{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除规格
	if err := tx.Delete(&spec).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetSpecificationsByProductID 获取商品的规格列表
func (s *SpecificationService) GetSpecificationsByProductID(productID int, merchantID int) ([]models.ProductSpecification, error) {
	// 检查商品是否属于该商户
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", productID, merchantID).First(&product)
	if result.Error != nil {
		return nil, errors.New("商品不存在或不属于该商户")
	}

	var specs []models.ProductSpecification
	result = s.DB.Where("product_id = ?", productID).Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).Order("sort ASC").Find(&specs)
	if result.Error != nil {
		return nil, result.Error
	}
	return specs, nil
}

// CreateSpecificationValue 创建规格值
func (s *SpecificationService) CreateSpecificationValue(value *models.ProductSpecificationValue, merchantID int) error {
	// 检查规格所属的商品是否属于该商户
	var spec models.ProductSpecification
	result := s.DB.First(&spec, value.SpecID)
	if result.Error != nil {
		return errors.New("规格不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", spec.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	result = s.DB.Create(value)
	return result.Error
}

// UpdateSpecificationValue 更新规格值
func (s *SpecificationService) UpdateSpecificationValue(valueID int, merchantID int, value string, image string, sortOrder int, status string) error {
	// 检查规格值所属的规格是否属于该商户
	var specValue models.ProductSpecificationValue
	result := s.DB.First(&specValue, valueID)
	if result.Error != nil {
		return errors.New("规格值不存在")
	}

	var spec models.ProductSpecification
	result = s.DB.First(&spec, specValue.SpecID)
	if result.Error != nil {
		return errors.New("规格不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", spec.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 更新规格值
	result = s.DB.Model(&specValue).Updates(map[string]interface{}{
		"value":  value,
		"image":  image,
		"sort":   sortOrder,
		"status": status,
	})
	return result.Error
}

// DeleteSpecificationValue 删除规格值
func (s *SpecificationService) DeleteSpecificationValue(valueID int, merchantID int) error {
	// 检查规格值所属的规格是否属于该商户
	var specValue models.ProductSpecificationValue
	result := s.DB.First(&specValue, valueID)
	if result.Error != nil {
		return errors.New("规格值不存在")
	}

	var spec models.ProductSpecification
	result = s.DB.First(&spec, specValue.SpecID)
	if result.Error != nil {
		return errors.New("规格不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", spec.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 检查规格值是否被SKU使用
	var count int64
	s.DB.Model(&models.ProductSkuSpec{}).Where("spec_value_id = ?", valueID).Count(&count)
	if count > 0 {
		return errors.New("规格值已被SKU使用，无法删除")
	}

	result = s.DB.Delete(&specValue)
	return result.Error
}

// CheckSpecificationUsed 检查规格是否被SKU使用
func (s *SpecificationService) CheckSpecificationUsed(specID int) bool {
	var count int64
	s.DB.Model(&models.ProductSkuSpec{}).Where("spec_id = ?", specID).Count(&count)
	return count > 0
}

// CheckSpecificationValueUsed 检查规格值是否被SKU使用
func (s *SpecificationService) CheckSpecificationValueUsed(valueID int) bool {
	var count int64
	s.DB.Model(&models.ProductSkuSpec{}).Where("spec_value_id = ?", valueID).Count(&count)
	return count > 0
}
