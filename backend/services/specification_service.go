package services

import (
	"errors"
	"goshopadmin/models"

	"gorm.io/gorm"
)

// ========== 响应结构体定义（精简版）==========

// SpecificationValueResp 规格值响应（精简版）
type SpecificationValueResp struct {
	ID     int    `json:"id"`
	Value  string `json:"value"`
	Sort   int    `json:"sort"`
	Status string `json:"status"`
}

// SpecificationResp 规格响应（精简版）
type SpecificationResp struct {
	ID     int                       `json:"id"`
	Name   string                    `json:"name"`
	Sort   int                       `json:"sort"`
	Values []SpecificationValueResp  `json:"values"`
}

// SpecificationService 规格服务
type SpecificationService struct {
	DB *gorm.DB
}

// NewSpecificationService 创建规格服务实例
func NewSpecificationService(db *gorm.DB) *SpecificationService {
	return &SpecificationService{DB: db}
}

// CreateSpecification 创建商品规格（单表插入）
func (s *SpecificationService) CreateSpecification(spec *models.ProductSpecification, merchantID int) error {
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", spec.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	result = s.DB.Exec(
		"INSERT INTO product_specifications (product_id, name, sort, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())",
		spec.ProductID,
		spec.Name,
		spec.Sort,
	)
	return result.Error
}

// UpdateSpecification 更新规格
func (s *SpecificationService) UpdateSpecification(specID int, merchantID int, name string, sortOrder int) error {
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

	result = s.DB.Model(&spec).Updates(map[string]interface{}{
		"name": name,
		"sort": sortOrder,
	})
	return result.Error
}

// DeleteSpecification 删除规格
func (s *SpecificationService) DeleteSpecification(specID int, merchantID int) error {
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

	var count int64
	s.DB.Model(&models.ProductSkuSpec{}).Where("spec_id = ?", specID).Count(&count)
	if count > 0 {
		return errors.New("规格已被SKU使用，无法删除")
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("spec_id = ?", specID).Delete(&models.ProductSpecificationValue{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&spec).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetSpecificationsByProductID 获取商品的规格列表（精简版）
func (s *SpecificationService) GetSpecificationsByProductID(
	productID int,
	merchantID int,
) ([]SpecificationResp, error) {
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", productID, merchantID).First(&product)
	if result.Error != nil {
		return nil, errors.New("商品不存在或不属于该商户")
	}

	var specs []models.ProductSpecification
	result = s.DB.Table("product_specifications").
		Select("id, product_id, name, sort").
		Where("product_id = ?", productID).
		Order("sort ASC").
		Find(&specs)
	if result.Error != nil {
		return nil, result.Error
	}

	if len(specs) == 0 {
		return []SpecificationResp{}, nil
	}

	specIDs := make([]int, len(specs))
	for i, spec := range specs {
		specIDs[i] = spec.ID
	}

	var specValues []models.ProductSpecificationValue
	result = s.DB.Table("product_specification_values").
		Select("id, spec_id, value, sort, status").
		Where("spec_id IN ?", specIDs).
		Order("sort ASC").
		Find(&specValues)
	if result.Error != nil {
		return nil, result.Error
	}

	valuesMap := make(map[int][]SpecificationValueResp)
	for _, v := range specValues {
		valuesMap[v.SpecID] = append(valuesMap[v.SpecID], SpecificationValueResp{
			ID:     v.ID,
			Value:  v.Value,
			Sort:   v.Sort,
			Status: v.Status,
		})
	}

	response := make([]SpecificationResp, len(specs))
	for i, spec := range specs {
		response[i] = SpecificationResp{
			ID:     spec.ID,
			Name:   spec.Name,
			Sort:   spec.Sort,
			Values: valuesMap[spec.ID],
		}
	}

	return response, nil
}

// CreateSpecificationValue 创建规格值（单表插入）
func (s *SpecificationService) CreateSpecificationValue(value *models.ProductSpecificationValue, merchantID int) error {
	var spec models.ProductSpecification
	result := s.DB.Select("id, product_id").First(&spec, value.SpecID)
	if result.Error != nil {
		return errors.New("规格不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", spec.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	result = s.DB.Exec(
		"INSERT INTO product_specification_values (spec_id, value, sort, status, created_at) VALUES (?, ?, ?, ?, NOW())",
		value.SpecID,
		value.Value,
		value.Sort,
		value.Status,
	)

	return result.Error
}

// UpdateSpecificationValue 更新规格值
func (s *SpecificationService) UpdateSpecificationValue(valueID int, merchantID int, value string, image string, sortOrder int, status string) error {
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
