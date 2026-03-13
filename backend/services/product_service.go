package services

import (
	"errors"
	"goshopadmin/constants"
	"goshopadmin/models"
	"strconv"

	"gorm.io/gorm"
)

// ProductService 商品服务
type ProductService struct {
	DB *gorm.DB
}

// NewProductService 创建商品服务实例
func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{DB: db}
}

// GetMerchantIDByUserID 根据用户ID获取商户ID
func (s *ProductService) GetMerchantIDByUserID(userID int) (int, error) {
	var merchantUser models.MerchantUser
	result := s.DB.Where("user_id = ?", userID).First(&merchantUser)
	if result.Error != nil {
		return 0, errors.New("用户不属于任何商户")
	}
	return merchantUser.MerchantID, nil
}

// GetProductsByMerchantID 获取商户的商品列表
func (s *ProductService) GetProductsByMerchantID(merchantID int) ([]models.Product, error) {
	var products []models.Product
	result := s.DB.Where("merchant_id = ?", merchantID).Preload("Category").Preload("Images").Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}
	return products, nil
}

// GetProductByID 根据ID获取商品详情
func (s *ProductService) GetProductByID(id int, merchantID int) (models.Product, error) {
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", id, merchantID).Preload("Category").Preload("Images").Preload("SKUs").First(&product)
	if result.Error != nil {
		return models.Product{}, result.Error
	}
	return product, nil
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(product *models.Product, merchantID int) error {
	// 开启事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	product.MerchantID = merchantID
	// 创建商品
	if err := tx.Create(product).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 自动创建默认SKU
	defaultSKU := models.ProductSKU{
		MerchantID: merchantID,
		ProductID:  product.ID,
		SKUCode:    "PROD-" + strconv.Itoa(product.ID) + "-DEFAULT",
		Attributes: "{\"type\": \"default\"}",
		Price:      product.Price,
		Stock:      product.Stock,
		Status:     "active",
	}

	if err := tx.Create(&defaultSKU).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(productID int, merchantID int, updateData map[string]interface{}) error {
	// 检查商品是否属于该商户
	var existingProduct models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", productID, merchantID).First(&existingProduct)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 使用map更新指定字段
	result = s.DB.Model(&existingProduct).Updates(updateData)
	return result.Error
}

// DeleteProduct 禁用商品
func (s *ProductService) DeleteProduct(id int, merchantID int) error {
	// 检查商品是否属于该商户
	var existingProduct models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&existingProduct)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 更新商品状态为禁用
	result = s.DB.Model(&existingProduct).Update("status", constants.StatusInactive)
	return result.Error
}

// GetCategoriesByMerchantID 获取商户的商品分类列表
func (s *ProductService) GetCategoriesByMerchantID(merchantID int) ([]models.ProductCategory, error) {
	var categories []models.ProductCategory
	result := s.DB.Where("merchant_id = ?", merchantID).Order("sort ASC").Find(&categories)
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}

// GetCategoryByID 根据ID获取商品分类详情
func (s *ProductService) GetCategoryByID(id int, merchantID int) (models.ProductCategory, error) {
	var category models.ProductCategory
	result := s.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&category)
	if result.Error != nil {
		return models.ProductCategory{}, result.Error
	}
	return category, nil
}

// CreateCategory 创建商品分类
func (s *ProductService) CreateCategory(category *models.ProductCategory, merchantID int) error {
	category.MerchantID = merchantID
	result := s.DB.Create(category)
	return result.Error
}

// UpdateCategory 更新商品分类
func (s *ProductService) UpdateCategory(category *models.ProductCategory, merchantID int) error {
	// 检查分类是否属于该商户
	var existingCategory models.ProductCategory
	result := s.DB.Where("id = ? AND merchant_id = ?", category.ID, merchantID).First(&existingCategory)
	if result.Error != nil {
		return errors.New("分类不存在或不属于该商户")
	}

	// 使用Updates方法更新，只更新非零值字段，避免覆盖时间戳
	result = s.DB.Model(&existingCategory).Updates(category)
	return result.Error
}

// DeleteCategory 禁用商品分类
func (s *ProductService) DeleteCategory(id int, merchantID int) error {
	// 检查分类是否属于该商户
	var existingCategory models.ProductCategory
	result := s.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&existingCategory)
	if result.Error != nil {
		return errors.New("分类不存在或不属于该商户")
	}

	// 检查分类下是否有商品
	var productCount int64
	s.DB.Model(&models.Product{}).Where("category_id = ?", id).Count(&productCount)
	if productCount > 0 {
		return errors.New("分类下存在商品，无法禁用")
	}

	// 更新分类状态为禁用
	result = s.DB.Model(&existingCategory).Update("status", constants.StatusInactive)
	return result.Error
}

// GetProductImageByID 根据ID获取商品图片
func (s *ProductService) GetProductImageByID(id int, merchantID int) (*models.ProductImage, error) {
	var image models.ProductImage
	result := s.DB.First(&image, id)
	if result.Error != nil {
		return nil, result.Error
	}

	// 检查图片所属的商品是否属于该商户
	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", image.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return nil, errors.New("商品不存在或不属于该商户")
	}

	return &image, nil
}

// AddProductImage 添加商品图片
func (s *ProductService) AddProductImage(image *models.ProductImage, merchantID int) error {
	// 检查商品是否属于该商户
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", image.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 如果是主图，将其他图片设置为非主图
	if image.IsMain {
		s.DB.Model(&models.ProductImage{}).Where("product_id = ?", image.ProductID).Update("is_main", false)
	}

	// 添加图片
	result = s.DB.Create(image)
	return result.Error
}

// DeleteProductImage 删除商品图片
func (s *ProductService) DeleteProductImage(id int, merchantID int) error {
	// 检查图片所属的商品是否属于该商户
	var image models.ProductImage
	result := s.DB.First(&image, id)
	if result.Error != nil {
		return errors.New("图片不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", image.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 删除图片
	result = s.DB.Delete(&image)
	return result.Error
}

// UpdateProductImage 更新商品图片
func (s *ProductService) UpdateProductImage(image *models.ProductImage, merchantID int) error {
	// 检查图片所属的商品是否属于该商户
	var existingImage models.ProductImage
	result := s.DB.First(&existingImage, image.ID)
	if result.Error != nil {
		return errors.New("图片不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", existingImage.ProductID, merchantID).First(&product)
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

	// 如果设置为主图，将其他图片设置为非主图
	if image.IsMain {
		if err := tx.Model(&models.ProductImage{}).Where("product_id = ? AND id != ?", existingImage.ProductID, image.ID).Update("is_main", false).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 更新图片信息
	if err := tx.Model(&existingImage).Updates(image).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// AddProductSKU 添加商品SKU
func (s *ProductService) AddProductSKU(sku *models.ProductSKU, merchantID int) error {
	// 检查商品是否属于该商户
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", sku.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	sku.MerchantID = merchantID

	// 开启事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 添加SKU
	if err := tx.Create(sku).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 重新计算商品总库存
	var totalStock int
	if err := tx.Model(&models.ProductSKU{}).Where("product_id = ? AND status = ?", sku.ProductID, "active").Select("COALESCE(SUM(stock), 0)").Scan(&totalStock).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新商品总库存
	if err := tx.Model(&product).Update("stock", totalStock).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// UpdateProductSKU 更新商品SKU
func (s *ProductService) UpdateProductSKU(sku *models.ProductSKU, merchantID int) error {
	// 检查SKU所属的商品是否属于该商户
	var existingSKU models.ProductSKU
	result := s.DB.First(&existingSKU, sku.ID)
	if result.Error != nil {
		return errors.New("SKU不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", existingSKU.ProductID, merchantID).First(&product)
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
	if err := tx.Model(&existingSKU).Updates(sku).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 重新计算商品总库存
	var totalStock int
	if err := tx.Model(&models.ProductSKU{}).Where("product_id = ? AND status = ?", existingSKU.ProductID, "active").Select("COALESCE(SUM(stock), 0)").Scan(&totalStock).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新商品总库存
	if err := tx.Model(&product).Update("stock", totalStock).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// DeleteProductSKU 禁用商品SKU
func (s *ProductService) DeleteProductSKU(id int, merchantID int) error {
	// 检查SKU所属的商品是否属于该商户
	var sku models.ProductSKU
	result := s.DB.First(&sku, id)
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
	s.DB.Model(&models.ProductSKU{}).Where("product_id = ? AND status = ?", sku.ProductID, "active").Count(&activeSKUCount)
	if activeSKUCount <= 1 {
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
	if err := tx.Model(&sku).Update("status", constants.StatusInactive).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 重新计算商品总库存
	var totalStock int
	if err := tx.Model(&models.ProductSKU{}).Where("product_id = ? AND status = ?", sku.ProductID, "active").Select("COALESCE(SUM(stock), 0)").Scan(&totalStock).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新商品总库存
	if err := tx.Model(&product).Update("stock", totalStock).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}
