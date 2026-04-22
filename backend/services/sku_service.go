package services

import (
	"errors"
	"fmt"
	"goshopadmin/models"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SkuService SKU服务
type SkuService struct {
	DB *gorm.DB
}

// NewSkuService 创建SKU服务实例
func NewSkuService(db *gorm.DB) *SkuService {
	return &SkuService{DB: db}
}

// CreateSku 创建单个SKU
func (s *SkuService) CreateSku(sku *models.ProductSku, specCombinations []models.ProductSkuSpec, merchantID int) error {
	// 检查商品是否属于该商户
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", sku.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	sku.MerchantID = merchantID

	// 确保 Attributes 字段有有效的 JSON 值
	if sku.Attributes == "" {
		sku.Attributes = "{}"
	}

	// 开启事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建SKU
	if err := tx.Create(sku).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 创建规格关联
	for i := range specCombinations {
		specCombinations[i].SkuID = sku.ID
		if err := tx.Create(&specCombinations[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 重新计算商品总库存
	var totalStock int
	err := tx.Model(&models.ProductSku{}).
		Where("product_id = ? AND status = ?", sku.ProductID, "active").
		Select("COALESCE(SUM(stock), 0)").
		Scan(&totalStock).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// 更新商品总库存
	if err := tx.Model(&product).Update("stock", totalStock).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// CreateActivitySku 创建活动专用SKU
func (s *SkuService) CreateActivitySku(sku *models.ProductSku, specCombinations []models.ProductSkuSpec, activityID int, merchantID int) error {
	// 检查活动是否存在
	var activity models.Activity
	result := s.DB.First(&activity, activityID)
	if result.Error != nil {
		return errors.New("活动不存在")
	}

	// 检查商品是否属于该商户
	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", sku.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 设置活动相关字段
	sku.MerchantID = merchantID
	sku.IsActivity = 1
	sku.ActivityID = activityID

	// 确保 Attributes 字段有有效的 JSON 值
	if sku.Attributes == "" {
		sku.Attributes = "{}"
	}

	// 开启事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建SKU
	if err := tx.Create(sku).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 创建规格关联
	for i := range specCombinations {
		specCombinations[i].SkuID = sku.ID
		if err := tx.Create(&specCombinations[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// BatchCreateSku 批量创建SKU
func (s *SkuService) BatchCreateSku(productID int, skus []models.ProductSku, specCombinations [][]models.ProductSkuSpec, merchantID int) error {
	// 检查商品是否属于该商户
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", productID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 开启事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 批量创建SKU
	for i := range skus {
		skus[i].ProductID = productID
		skus[i].MerchantID = merchantID
		if skus[i].SkuCode == "" {
			skus[i].SkuCode = fmt.Sprintf("PROD-%d-%d", productID, i+1)
		}
		// 确保 Attributes 字段有有效的 JSON 值
		if skus[i].Attributes == "" {
			skus[i].Attributes = "{}"
		}
		if err := tx.Create(&skus[i]).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 创建规格关联
		if i < len(specCombinations) {
			for j := range specCombinations[i] {
				specCombinations[i][j].SkuID = skus[i].ID
				if err := tx.Create(&specCombinations[i][j]).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	// 重新计算商品总库存
	var totalStock int
	err := tx.Model(&models.ProductSku{}).
		Where("product_id = ? AND status = ?", productID, "active").
		Select("COALESCE(SUM(stock), 0)").
		Scan(&totalStock).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// 更新商品总库存
	if err := tx.Model(&product).Update("stock", totalStock).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UpdateSku 更新SKU
func (s *SkuService) UpdateSku(skuID int, updates map[string]interface{}, specCombinations []models.ProductSkuSpec, merchantID int) error {
	// 检查SKU所属的商品是否属于该商户
	var existingSku models.ProductSku
	result := s.DB.First(&existingSku, skuID)
	if result.Error != nil {
		return errors.New("SKU不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", existingSku.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 开启事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新SKU
	if len(updates) > 0 {
		if err := tx.Model(&existingSku).Updates(updates).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 如果有规格组合更新，先删除旧的，再创建新的
	if len(specCombinations) > 0 {
		if err := tx.Where("sku_id = ?", skuID).Delete(&models.ProductSkuSpec{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		for i := range specCombinations {
			specCombinations[i].SkuID = skuID
			if err := tx.Create(&specCombinations[i]).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// 重新计算商品总库存
	var totalStock int
	err := tx.Model(&models.ProductSku{}).
		Where("product_id = ? AND status = ?", existingSku.ProductID, "active").
		Select("COALESCE(SUM(stock), 0)").
		Scan(&totalStock).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// 更新商品总库存
	if err := tx.Model(&product).Update("stock", totalStock).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// DeleteSku 删除SKU（禁用）
func (s *SkuService) DeleteSku(skuID int, merchantID int) error {
	// 检查SKU所属的商品是否属于该商户
	var sku models.ProductSku
	result := s.DB.First(&sku, skuID)
	if result.Error != nil {
		return errors.New("SKU不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", sku.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 检查是否是最后一个SKU
	var activeSKUCount int64
	s.DB.Model(&models.ProductSku{}).
		Where("product_id = ? AND status = ? AND id != ?", sku.ProductID, "active", skuID).
		Count(&activeSKUCount)
	if activeSKUCount == 0 {
		return errors.New("不能删除最后一个SKU")
	}

	// 开启事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新SKU状态为禁用
	if err := tx.Model(&sku).Update("status", "inactive").Error; err != nil {
		tx.Rollback()
		return err
	}

	// 重新计算商品总库存
	var totalStock int
	err := tx.Model(&models.ProductSku{}).
		Where("product_id = ? AND status = ?", sku.ProductID, "active").
		Select("COALESCE(SUM(stock), 0)").
		Scan(&totalStock).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// 更新商品总库存
	if err := tx.Model(&product).Update("stock", totalStock).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetSkusByProductID 获取商品的SKU列表
func (s *SkuService) GetSkusByProductID(productID int, merchantID int) ([]models.ProductSku, error) {
	// 检查商品是否属于该商户
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", productID, merchantID).First(&product)
	if result.Error != nil {
		return nil, errors.New("商品不存在或不属于该商户")
	}

	var skus []models.ProductSku
	result = s.DB.Where("product_id = ?", productID).Preload("Specs", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Spec").Preload("SpecValue")
	}).Find(&skus)
	if result.Error != nil {
		return nil, result.Error
	}
	return skus, nil
}

// GetSkuByID 根据ID获取SKU详情
func (s *SkuService) GetSkuByID(skuID int, merchantID int) (models.ProductSku, error) {
	var sku models.ProductSku
	result := s.DB.First(&sku, skuID)
	if result.Error != nil {
		return models.ProductSku{}, errors.New("SKU不存在")
	}

	// 检查SKU所属的商品是否属于该商户
	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", sku.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return models.ProductSku{}, errors.New("商品不存在或不属于该商户")
	}

	// 加载规格关联
	s.DB.Model(&sku).Association("Specs").Find(&sku.Specs)
	return sku, nil
}

// SkuWithSpecCombinations 包含规格组合的SKU
type SkuWithSpecCombinations struct {
	models.ProductSku
	SpecCombinations []models.ProductSkuSpec `json:"spec_combinations"`
}

// GenerateSkusFromSpecs 根据规格组合自动生成SKU
func (s *SkuService) GenerateSkusFromSpecs(productID int, basePrice float64, merchantID int) ([]SkuWithSpecCombinations, error) {
	// 检查商品是否属于该商户
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", productID, merchantID).First(&product)
	if result.Error != nil {
		return nil, errors.New("商品不存在或不属于该商户")
	}

	// 获取商品的所有规格和规格值
	var specs []models.ProductSpecification
	result = s.DB.Where("product_id = ?", productID).Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", "active").Order("sort ASC")
	}).Order("sort ASC").Find(&specs)
	if result.Error != nil {
		return nil, result.Error
	}

	if len(specs) == 0 {
		return nil, errors.New("商品没有配置规格")
	}

	// 检查是否所有规格都有规格值
	for _, spec := range specs {
		if len(spec.Values) == 0 {
			return nil, errors.New("规格 '" + spec.Name + "' 没有配置规格值")
		}
	}

	// 生成规格组合
	var combinations [][]models.ProductSpecificationValue
	var generateCombinations func(specIndex int, current []models.ProductSpecificationValue)
	generateCombinations = func(specIndex int, current []models.ProductSpecificationValue) {
		if specIndex == len(specs) {
			// 复制当前组合
			combo := make([]models.ProductSpecificationValue, len(current))
			copy(combo, current)
			combinations = append(combinations, combo)
			return
		}

		for _, value := range specs[specIndex].Values {
			current = append(current, value)
			generateCombinations(specIndex+1, current)
			current = current[:len(current)-1]
		}
	}
	generateCombinations(0, []models.ProductSpecificationValue{})

	// 获取已存在的SKU编码
	var existingSkus []models.ProductSku
	if err := s.DB.Where("product_id = ?", productID).Select("sku_code").Find(&existingSkus).Error; err != nil {
		return nil, err
	}

	// 创建已存在SKU编码的映射，用于快速查找
	existingSkuMap := make(map[string]bool)
	for _, sku := range existingSkus {
		existingSkuMap[sku.SkuCode] = true
	}

	// 生成SKU列表，过滤掉已存在的SKU
	var skus []SkuWithSpecCombinations
	for _, combo := range combinations {
		// 生成SKU编码
		var skuCodeParts []string
		skuCodeParts = append(skuCodeParts, "PROD-"+strconv.Itoa(productID))
		for _, value := range combo {
			skuCodeParts = append(skuCodeParts, value.Value)
		}
		skuCode := strings.Join(skuCodeParts, "-")

		// 检查SKU编码是否已存在，如果存在则跳过
		if existingSkuMap[skuCode] {
			continue
		}

		sku := models.ProductSku{
			ProductID:     productID,
			MerchantID:    merchantID,
			SkuCode:       skuCode,
			Price:         decimal.NewFromFloat(basePrice),
			OriginalPrice: decimal.NewFromInt(0),
			Stock:         0,
			Status:        "active",
		}

		// 生成规格组合关联（用于返回给前端展示）
		var specCombos []models.ProductSkuSpec
		for _, value := range combo {
			specCombos = append(specCombos, models.ProductSkuSpec{
				SpecID:      value.SpecID,
				SpecValueID: value.ID,
			})
		}

		skus = append(skus, SkuWithSpecCombinations{
			ProductSku:       sku,
			SpecCombinations: specCombos,
		})
	}

	return skus, nil
}
