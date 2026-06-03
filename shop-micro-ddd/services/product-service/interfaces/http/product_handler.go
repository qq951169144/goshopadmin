package http

import (
	"net/http"
	"strconv"

	"product-service/domain/product"

	"github.com/gin-gonic/gin"
)

// ProductHandler 商品 HTTP Handler
type ProductHandler struct {
	productRepo product.ProductRepository
}

// NewProductHandler 创建商品 Handler
func NewProductHandler(repo product.ProductRepository) *ProductHandler {
	return &ProductHandler{productRepo: repo}
}

// GetProductAndSKU 获取商品和SKU信息（内部 API，供其他服务调用）
// GET /api/internal/products/:productID/skus/:skuID
func (h *ProductHandler) GetProductAndSKU(ctx *gin.Context) {
	productID, err := strconv.Atoi(ctx.Param("productID"))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4001, "message": "商品ID无效", "data": nil})
		return
	}
	skuID, err := strconv.Atoi(ctx.Param("skuID"))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4001, "message": "SKU ID无效", "data": nil})
		return
	}

	// 查询商品
	p, err := h.productRepo.FindByID(ctx.Request.Context(), productID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4042, "message": "商品不存在", "data": nil})
		return
	}

	// 查找指定SKU
	sku, found := p.FindSKU(skuID)
	if !found {
		ctx.JSON(http.StatusOK, gin.H{"code": 4042, "message": "SKU不存在", "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"name":      p.Name(),
			"sku_attrs": sku.Attrs(),
			"price":     sku.Price(),
			"stock":     sku.Stock(),
		},
	})
}

// deductStockRequest 扣减库存请求
type deductStockRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// DeductStock 扣减库存（内部 API）
// POST /api/internal/skus/:skuID/deduct
func (h *ProductHandler) DeductStock(ctx *gin.Context) {
	skuID, err := strconv.Atoi(ctx.Param("skuID"))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4001, "message": "SKU ID无效", "data": nil})
		return
	}

	var req deductStockRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4001, "message": "参数错误", "data": nil})
		return
	}

	// 查询SKU
	sku, err := h.productRepo.FindSKUByID(ctx.Request.Context(), skuID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4042, "message": "SKU不存在", "data": nil})
		return
	}

	// 扣减库存（充血模型方法）
	if err := sku.DeductStock(req.Quantity); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4092, "message": err.Error(), "data": nil})
		return
	}

	// 持久化
	if err := h.productRepo.SaveSKU(ctx.Request.Context(), sku); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 5001, "message": "保存失败", "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// restoreStockRequest 恢复库存请求
type restoreStockRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// RestoreStock 恢复库存（内部 API）
// POST /api/internal/skus/:skuID/restore
func (h *ProductHandler) RestoreStock(ctx *gin.Context) {
	skuID, err := strconv.Atoi(ctx.Param("skuID"))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4001, "message": "SKU ID无效", "data": nil})
		return
	}

	var req restoreStockRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4001, "message": "参数错误", "data": nil})
		return
	}

	// 查询SKU
	sku, err := h.productRepo.FindSKUByID(ctx.Request.Context(), skuID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4042, "message": "SKU不存在", "data": nil})
		return
	}

	// 恢复库存（充血模型方法）
	sku.RestoreStock(req.Quantity)

	// 持久化
	if err := h.productRepo.SaveSKU(ctx.Request.Context(), sku); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 5001, "message": "保存失败", "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}
