package services

import (
	"errors"
	"shop-backend/constants"
	"shop-backend/models"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// SpecificationService 规格服务
type SpecificationService struct {
	db *gorm.DB
}

// NewSpecificationService 创建规格服务实例
func NewSpecificationService(db *gorm.DB) *SpecificationService {
	return &SpecificationService{db: db}
}

// SpecificationInfo 规格信息结构
type SpecificationInfo struct {
	ID     int                      `json:"id"`
	Name   string                   `json:"name"`
	Values []SpecificationValueInfo `json:"values"`
}

// SpecificationValueInfo 规格值信息结构
type SpecificationValueInfo struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
	Image string `json:"image"`
}

// SKUInfoWithSpecs 带规格信息的SKU
type SKUInfoWithSpecs struct {
	ID              int            `json:"id"`
	SKUCode         string         `json:"sku_code"`
	Price           float64        `json:"price"`
	OriginalPrice   float64        `json:"original_price"`
	Stock           int            `json:"stock"`
	Status          string         `json:"status"`
	SpecCombination map[string]int `json:"spec_combination"` // spec_id -> spec_value_id
}

// ProductDetailWithSpecs 带规格信息的商品详情
type ProductDetailWithSpecs struct {
	ID             int                 `json:"id"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	Detail         string              `json:"detail"`
	Price          float64             `json:"price"`
	Image          string              `json:"image"`
	Images         []string            `json:"images"`
	Specifications []SpecificationInfo `json:"specifications"`
	SKUs           []SKUInfoWithSpecs  `json:"sku_list"`
	PriceRange     PriceRange          `json:"price_range"`
	Sales          int                 `json:"sales"`
	ReviewsCount   int                 `json:"reviews_count"`
}

// PriceRange 价格范围
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// GetProductDetailWithSpecs 获取带规格信息的商品详情
func (s *SpecificationService) GetProductDetailWithSpecs(productID int) (*ProductDetailWithSpecs, error) {
	var product models.Product
	result := s.db.First(&product, productID)
	if result.RowsAffected == 0 {
		return nil, errors.New("商品不存在")
	}

	// 查询商品图片列表
	var productImages []models.ProductImage
	s.db.Where("product_id = ?", productID).Order("is_main DESC, sort ASC").Find(&productImages)

	var images []string
	var mainImage string
	for _, img := range productImages {
		if img.ImageURL != "" {
			images = append(images, img.ImageURL)
			if img.IsMain && mainImage == "" {
				mainImage = img.ImageURL
			}
		}
	}
	if mainImage == "" && len(images) > 0 {
		mainImage = images[0]
	}

	// 查询规格列表
	var specifications []models.ProductSpecification
	s.db.Where("product_id = ?", productID).Order("sort ASC").Find(&specifications)

	var specInfos []SpecificationInfo
	for _, spec := range specifications {
		var values []models.ProductSpecificationValue
		s.db.Where("spec_id = ? AND status = ?", spec.ID, constants.StatusActive).Order("sort ASC").Find(&values)

		var valueInfos []SpecificationValueInfo
		for _, value := range values {
			valueInfos = append(valueInfos, SpecificationValueInfo{
				ID:    value.ID,
				Value: value.Value,
			})
		}

		specInfos = append(specInfos, SpecificationInfo{
			ID:     spec.ID,
			Name:   spec.Name,
			Values: valueInfos,
		})
	}

	// 查询SKU列表
	var skus []models.ProductSKU
	s.db.Where("product_id = ?", productID).Find(&skus)

	var skuInfos []SKUInfoWithSpecs
	var minPrice, maxPrice float64
	for i, sku := range skus {
		// 查询SKU的规格组合
		var skuSpecs []models.ProductSKUSpec
		s.db.Where("sku_id = ?", sku.ID).Find(&skuSpecs)

		specCombination := make(map[string]int)
		for _, skuSpec := range skuSpecs {
			specCombination[strconv.Itoa(skuSpec.SpecID)] = skuSpec.SpecValueID
		}

		skuInfos = append(skuInfos, SKUInfoWithSpecs{
			ID:              sku.ID,
			SKUCode:         sku.SKUCode,
			Price:           sku.Price,
			OriginalPrice:   sku.OriginalPrice,
			Stock:           sku.Stock,
			Status:          sku.Status,
			SpecCombination: specCombination,
		})

		// 计算价格范围
		if i == 0 || sku.Price < minPrice {
			minPrice = sku.Price
		}
		if i == 0 || sku.Price > maxPrice {
			maxPrice = sku.Price
		}
	}

	// 如果没有SKU，使用商品默认价格
	if len(skuInfos) == 0 {
		skuInfos = append(skuInfos, SKUInfoWithSpecs{
			ID:              0,
			SKUCode:         "default",
			Price:           product.Price,
			OriginalPrice:   0,
			Stock:           product.Stock,
			Status:          "active",
			SpecCombination: make(map[string]int),
		})
		minPrice = product.Price
		maxPrice = product.Price
	}

	return &ProductDetailWithSpecs{
		ID:             product.ID,
		Name:           product.Name,
		Description:    product.Description,
		Detail:         product.Detail,
		Price:          product.Price,
		Image:          mainImage,
		Images:         images,
		Specifications: specInfos,
		SKUs:           skuInfos,
		PriceRange: PriceRange{
			Min: minPrice,
			Max: maxPrice,
		},
		Sales:        0,
		ReviewsCount: 0,
	}, nil
}

// GetSKUsByProductID 获取商品的SKU列表
func (s *SpecificationService) GetSKUsByProductID(productID int) ([]SKUInfoWithSpecs, error) {
	var skus []models.ProductSKU
	s.db.Where("product_id = ?", productID).Find(&skus)

	var skuInfos []SKUInfoWithSpecs
	for _, sku := range skus {
		// 查询SKU的规格组合
		var skuSpecs []models.ProductSKUSpec
		s.db.Where("sku_id = ?", sku.ID).Find(&skuSpecs)

		specCombination := make(map[string]int)
		for _, skuSpec := range skuSpecs {
			specCombination[strconv.Itoa(skuSpec.SpecID)] = skuSpec.SpecValueID
		}

		skuInfos = append(skuInfos, SKUInfoWithSpecs{
			ID:              sku.ID,
			SKUCode:         sku.SKUCode,
			Price:           sku.Price,
			OriginalPrice:   sku.OriginalPrice,
			Stock:           sku.Stock,
			Status:          sku.Status,
			SpecCombination: specCombination,
		})
	}

	return skuInfos, nil
}

// GetSKUBySpecCombination 根据规格组合查询SKU
func (s *SpecificationService) GetSKUBySpecCombination(productID int, specQuery string) (*SKUInfoWithSpecs, error) {
	// 解析规格查询参数，格式: "1:1,2:4" 表示 spec_id=1 对应 spec_value_id=1, spec_id=2 对应 spec_value_id=4
	specPairs := strings.Split(specQuery, ",")
	if len(specPairs) == 0 {
		return nil, errors.New("规格参数格式错误")
	}

	// 构建规格条件
	specConditions := make(map[int]int)
	for _, pair := range specPairs {
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			continue
		}
		specID, _ := strconv.Atoi(parts[0])
		specValueID, _ := strconv.Atoi(parts[1])
		if specID > 0 && specValueID > 0 {
			specConditions[specID] = specValueID
		}
	}

	if len(specConditions) == 0 {
		return nil, errors.New("规格参数格式错误")
	}

	// 查询所有SKU
	var skus []models.ProductSKU
	s.db.Where("product_id = ?", productID).Find(&skus)

	// 查找匹配的SKU
	for _, sku := range skus {
		var skuSpecs []models.ProductSKUSpec
		s.db.Where("sku_id = ?", sku.ID).Find(&skuSpecs)

		// 检查是否匹配所有规格条件
		match := true
		for specID, specValueID := range specConditions {
			found := false
			for _, skuSpec := range skuSpecs {
				if skuSpec.SpecID == specID && skuSpec.SpecValueID == specValueID {
					found = true
					break
				}
			}
			if !found {
				match = false
				break
			}
		}

		if match {
			specCombination := make(map[string]int)
			for _, skuSpec := range skuSpecs {
				specCombination[strconv.Itoa(skuSpec.SpecID)] = skuSpec.SpecValueID
			}

			return &SKUInfoWithSpecs{
				ID:              sku.ID,
				SKUCode:         sku.SKUCode,
				Price:           sku.Price,
				OriginalPrice:   sku.OriginalPrice,
				Stock:           sku.Stock,
				Status:          sku.Status,
				SpecCombination: specCombination,
			}, nil
		}
	}

	return nil, errors.New("未找到匹配的SKU")
}
