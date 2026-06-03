package http

import (
	"net/http"
	"strconv"

	"customer-service/domain/customer"

	"github.com/gin-gonic/gin"
)

// CustomerHandler 客户 HTTP Handler
type CustomerHandler struct {
	customerRepo customer.CustomerRepository
}

// NewCustomerHandler 创建客户 Handler
func NewCustomerHandler(repo customer.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{customerRepo: repo}
}

// VerifyAddress 验证地址（内部 API，供其他服务调用）
// GET /api/internal/customers/:customerID/addresses/:addressID/verify
func (h *CustomerHandler) VerifyAddress(ctx *gin.Context) {
	customerID, err := strconv.Atoi(ctx.Param("customerID"))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4001, "message": "客户ID无效", "data": nil})
		return
	}
	addressID, err := strconv.Atoi(ctx.Param("addressID"))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4001, "message": "地址ID无效", "data": nil})
		return
	}

	// 查询客户
	c, err := h.customerRepo.FindByID(ctx.Request.Context(), customerID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4041, "message": "客户不存在", "data": nil})
		return
	}

	// 验证地址
	merchantID, err := c.VerifyAddress(addressID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4031, "message": err.Error(), "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"merchant_id": merchantID,
		},
	})
}
