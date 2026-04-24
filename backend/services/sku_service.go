package services

import (
	"encoding/json"
	"errors"
	"goshopadmin/models"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// ========== 响应结构体定义（精简版）==========

// SkuSpecComboResp SKU规格组合响应（精简版）
type SkuSpecComboResp struct {
	SpecID      int `json:"spec_id"`
	SpecValueID int `json:"spec_value_id"`
}

// SkuResp SKU响应（精简版）
type SkuResp struct {
	ID               int                `json:"id"`
	SkuCode          string             `json:"sku_code"`
	Price            float64            `json:"price"`
	OriginalPrice    float64            `json:"original_price"`
	Stock            int                `json:"stock"`
	Status           string             `json:"status"`
	IsActivity       bool               `json:"is_activity"`
	ActivityID       int                `json:"activity_id"`
	SpecCombinations []SkuSpecComboResp `json:"spec_combinations"`
}

// SkuPreviewResp SKU预览响应（精简版）
type SkuPreviewResp struct {
	SkuCode          string             `json:"sku_code"`
	Price            float64            `json:"price"`
	SpecCombinations []SkuSpecComboResp `json:"spec_combinations"`
}

// SkuService SKU服务
type SkuService struct {
	DB *gorm.DB
}

// NewSkuService 创建SKU服务实例
func NewSkuService(db *gorm.DB) *SkuService {
	return &SkuService{DB: db}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i != 0
}

// CreateSku 创建单个SKU（单表插入）
func (s *SkuService) CreateSku(sku *models.ProductSku, specCombinations []models.ProductSkuSpec, merchantID int) error {
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", sku.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	sku.MerchantID = merchantID

	// 收集规格组合数据，用于后续插入
	if sku.Attributes == "" {
		attrs := make(map[string]string)
		for _, spec := range specCombinations {
			// 查询规格名称
			var specification models.ProductSpecification
			s.DB.Where("id = ?", spec.SpecID).Select("id, name").First(&specification)
			// 查询规格值
			var specValue models.ProductSpecificationValue
			s.DB.Where("id = ?", spec.SpecValueID).Select("id, value").First(&specValue)
			if specification.ID > 0 && specValue.ID > 0 {
				attrs[specification.Name] = specValue.Value
			}
		}
		if len(attrs) > 0 {
			attrsJSON, _ := json.Marshal(attrs)
			sku.Attributes = string(attrsJSON)
		}
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 使用 GORM Create 方法插入 SKU
	if err := tx.Create(sku).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 批量插入规格组合数据
	var skuSpecs []models.ProductSkuSpec
	for _, combo := range specCombinations {
		skuSpecs = append(skuSpecs, models.ProductSkuSpec{
			SkuID:       sku.ID,
			SpecID:      combo.SpecID,
			SpecValueID: combo.SpecValueID,
		})
	}
	if len(skuSpecs) > 0 {
		if err := tx.CreateInBatches(skuSpecs, 100).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	var totalStock int
	err := tx.Model(&models.ProductSku{}).
		Where("product_id = ? AND status = ?", sku.ProductID, "active").
		Select("COALESCE(SUM(stock), 0)").
		Scan(&totalStock).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&product).Update("stock", totalStock).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// BatchCreateSku 批量创建SKU（单表插入）
func (s *SkuService) BatchCreateSku(productID int, skus []models.ProductSku, specCombinations [][]models.ProductSkuSpec, merchantID int) error {
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", productID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i := range skus {
		skus[i].ProductID = productID
		skus[i].MerchantID = merchantID

		// if skus[i].SkuCode == "" {
		// 	// TODO:这里其实可以优化，这里和前端的生成规则对应不上，需要优化或者删除
		// 	skus[i].SkuCode = fmt.Sprintf("PROD-%d-%d", productID, i+1)
		// }

		// 收集规格组合数据，用于后续插入
		var skuSpecs []models.ProductSkuSpec
		if skus[i].Attributes == "" {
			attrs := make(map[string]string)
			for _, spec := range specCombinations[i] {
				// 查询规格名称
				var specification models.ProductSpecification
				tx.Where("id = ?", spec.SpecID).Select("id, name").First(&specification)
				// 查询规格值
				var specValue models.ProductSpecificationValue
				tx.Where("id = ?", spec.SpecValueID).Select("id, value").First(&specValue)
				if specification.ID > 0 && specValue.ID > 0 {
					attrs[specification.Name] = specValue.Value
				}
				// 收集规格组合数据
				skuSpecs = append(skuSpecs, models.ProductSkuSpec{
					SpecID:      spec.SpecID,
					SpecValueID: spec.SpecValueID,
				})
			}
			if len(attrs) > 0 {
				attrsJSON, _ := json.Marshal(attrs)
				skus[i].Attributes = string(attrsJSON)
			}
		} else {
			// 如果已有 Attributes，仍然需要收集规格组合数据
			for _, spec := range specCombinations[i] {
				skuSpecs = append(skuSpecs, models.ProductSkuSpec{
					SpecID:      spec.SpecID,
					SpecValueID: spec.SpecValueID,
				})
			}
		}

		// 使用 GORM Create 方法插入 SKU
		err := tx.Omit("ProductSkuSpec", "Product").Create(&skus[i]).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		// 批量插入规格组合数据
		for j := range skuSpecs {
			skuSpecs[j].SkuID = skus[i].ID
		}
		if len(skuSpecs) > 0 {
			err := tx.Omit("ProductSku", "ProductSpecification", "ProductSpecificationValue").CreateInBatches(skuSpecs, 100).Error
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	var totalStock int
	err := tx.Model(&models.ProductSku{}).
		Where("product_id = ? AND status = ?", productID, "active").
		Select("COALESCE(SUM(stock), 0)").
		Scan(&totalStock).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&product).Update("stock", totalStock).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UpdateSku 更新SKU
func (s *SkuService) UpdateSku(skuID int, updates map[string]interface{}, specCombinations []models.ProductSkuSpec, merchantID int) error {
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

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if len(updates) > 0 {
		if err := tx.Model(&existingSku).Updates(updates).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

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

	var totalStock int
	err := tx.Model(&models.ProductSku{}).
		Where("product_id = ? AND status = ?", existingSku.ProductID, "active").
		Select("COALESCE(SUM(stock), 0)").
		Scan(&totalStock).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&product).Update("stock", totalStock).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// DeleteSku 删除SKU（禁用）
func (s *SkuService) DeleteSku(skuID int, merchantID int) error {
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

	var activeSKUCount int64
	s.DB.Model(&models.ProductSku{}).
		Where("product_id = ? AND status = ? AND id != ?", sku.ProductID, "active", skuID).
		Count(&activeSKUCount)
	if activeSKUCount == 0 {
		return errors.New("不能删除最后一个SKU")
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&sku).Update("status", "inactive").Error; err != nil {
		tx.Rollback()
		return err
	}

	var totalStock int
	err := tx.Model(&models.ProductSku{}).
		Where("product_id = ? AND status = ?", sku.ProductID, "active").
		Select("COALESCE(SUM(stock), 0)").
		Scan(&totalStock).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&product).Update("stock", totalStock).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetSkusByProductID 获取商品的SKU列表（精简版）
func (s *SkuService) GetSkusByProductID(
	productID int,
	merchantID int,
) ([]SkuResp, error) {
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", productID, merchantID).First(&product)
	if result.Error != nil {
		return nil, errors.New("商品不存在或不属于该商户")
	}

	var skus []models.ProductSku
	result = s.DB.Table("product_skus").
		Select("id, product_id, sku_code, price, original_price, stock, status, is_activity, activity_id").
		Where("product_id = ?", productID).
		Find(&skus)
	if result.Error != nil {
		return nil, result.Error
	}

	if len(skus) == 0 {
		return []SkuResp{}, nil
	}

	skuIDs := make([]int, len(skus))
	for i, sku := range skus {
		skuIDs[i] = sku.ID
	}

	var skuSpecs []models.ProductSkuSpec
	result = s.DB.Table("product_sku_specs").
		Select("sku_id, spec_id, spec_value_id").
		Where("sku_id IN ?", skuIDs).
		Find(&skuSpecs)
	if result.Error != nil {
		return nil, result.Error
	}

	specsMap := make(map[int][]SkuSpecComboResp)
	for _, spec := range skuSpecs {
		specsMap[spec.SkuID] = append(specsMap[spec.SkuID], SkuSpecComboResp{
			SpecID:      spec.SpecID,
			SpecValueID: spec.SpecValueID,
		})
	}

	response := make([]SkuResp, len(skus))
	for i, sku := range skus {
		response[i] = SkuResp{
			ID:               sku.ID,
			SkuCode:          sku.SkuCode,
			Price:            sku.Price.InexactFloat64(),
			OriginalPrice:    sku.OriginalPrice.InexactFloat64(),
			Stock:            sku.Stock,
			Status:           sku.Status,
			IsActivity:       intToBool(sku.IsActivity),
			ActivityID:       sku.ActivityID,
			SpecCombinations: specsMap[sku.ID],
		}
	}

	return response, nil
}

// GetSkuByID 根据ID获取SKU详情
func (s *SkuService) GetSkuByID(skuID int, merchantID int) (models.ProductSku, error) {
	var sku models.ProductSku
	result := s.DB.First(&sku, skuID)
	if result.Error != nil {
		return models.ProductSku{}, errors.New("SKU不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", sku.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return models.ProductSku{}, errors.New("商品不存在或不属于该商户")
	}

	s.DB.Model(&sku).Association("Specs").Find(&sku.Specs)
	return sku, nil
}

// GenerateSkusFromSpecs 根据规格组合自动生成SKU预览（精简版）
func (s *SkuService) GenerateSkusFromSpecs(
	productID int,
	basePrice float64,
	merchantID int,
) ([]SkuPreviewResp, error) {
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
		return nil, errors.New("商品没有配置规格")
	}

	specIDs := make([]int, len(specs))
	for i, spec := range specs {
		specIDs[i] = spec.ID
	}

	var specValues []models.ProductSpecificationValue
	result = s.DB.Table("product_specification_values").
		Select("id, spec_id, value, sort").
		Where("spec_id IN ? AND status = ?", specIDs, "active").
		Order("sort ASC").
		Find(&specValues)
	if result.Error != nil {
		return nil, result.Error
	}

	valuesMap := make(map[int][]models.ProductSpecificationValue)
	for _, v := range specValues {
		valuesMap[v.SpecID] = append(valuesMap[v.SpecID], v)
	}

	for _, spec := range specs {
		if len(valuesMap[spec.ID]) == 0 {
			return nil, errors.New("规格 '" + spec.Name + "' 没有配置规格值")
		}
	}

	var combinations [][]models.ProductSpecificationValue
	var generateCombinations func(specIndex int, current []models.ProductSpecificationValue)

	generateCombinations = func(specIndex int, current []models.ProductSpecificationValue) {
		if specIndex == len(specs) {
			combo := make([]models.ProductSpecificationValue, len(current))
			copy(combo, current)
			combinations = append(combinations, combo)
			return
		}

		for _, value := range valuesMap[specs[specIndex].ID] {
			current = append(current, value)
			generateCombinations(specIndex+1, current)
			current = current[:len(current)-1]
		}
	}
	generateCombinations(0, []models.ProductSpecificationValue{})

	var existingSkus []models.ProductSku
	err := s.DB.Table("product_skus").
		Select("sku_code").
		Where("product_id = ?", productID).
		Find(&existingSkus).Error
	if err != nil {
		return nil, err
	}

	existingSkuMap := make(map[string]bool)
	for _, sku := range existingSkus {
		existingSkuMap[sku.SkuCode] = true
	}

	var preview []SkuPreviewResp
	for _, combo := range combinations {
		var skuCodeParts []string
		skuCodeParts = append(skuCodeParts, "PROD-"+strconv.Itoa(productID))

		specCombinations := make([]SkuSpecComboResp, len(combo))
		for j, value := range combo {
			skuCodeParts = append(skuCodeParts, value.Value)
			specCombinations[j] = SkuSpecComboResp{
				SpecID:      value.SpecID,
				SpecValueID: value.ID,
			}
		}
		skuCode := strings.Join(skuCodeParts, "-")

		if existingSkuMap[skuCode] {
			continue
		}

		preview = append(preview, SkuPreviewResp{
			SkuCode:          skuCode,
			Price:            basePrice,
			SpecCombinations: specCombinations,
		})
	}

	return preview, nil
}
