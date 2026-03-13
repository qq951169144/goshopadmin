package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"shop-backend/errors"
	"shop-backend/services"
)

// AddressController 地址控制器
type AddressController struct {
	BaseController
	addressService *services.AddressService
}

// NewAddressController 创建地址控制器实例
func NewAddressController(addressService *services.AddressService) *AddressController {
	return &AddressController{
		addressService: addressService,
	}
}

// GetAddresses 获取地址列表
func (c *AddressController) GetAddresses(ctx *gin.Context) {
	// 从上下文中获取用户ID
	customerID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	addresses, err := c.addressService.GetAddressList(customerID.(uint))
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"addresses": addresses,
	})
}

// GetAddress 获取单个地址
func (c *AddressController) GetAddress(ctx *gin.Context) {
	// 从上下文中获取用户ID
	customerID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	address, err := c.addressService.GetAddressByID(customerID.(uint), uint(id))
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}
	if address == nil {
		c.ResponseError(ctx, errors.CodeNotFound, nil)
		return
	}

	c.ResponseSuccess(ctx, address)
}

// CreateAddressRequest 创建地址请求结构
type CreateAddressRequest struct {
	Name          string `json:"name" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	Province      string `json:"province" binding:"required"`
	City          string `json:"city" binding:"required"`
	District      string `json:"district" binding:"required"`
	DetailAddress string `json:"detail_address" binding:"required"`
	IsDefault     bool   `json:"is_default"`
}

// CreateAddress 创建地址
func (c *AddressController) CreateAddress(ctx *gin.Context) {
	// 从上下文中获取用户ID
	customerID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	var req CreateAddressRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	address, err := c.addressService.CreateAddress(customerID.(uint), services.CreateAddressRequest{
		Name:          req.Name,
		Phone:         req.Phone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		DetailAddress: req.DetailAddress,
		IsDefault:     req.IsDefault,
	})
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "地址创建成功",
		"address": address,
	})
}

// UpdateAddressRequest 更新地址请求结构
type UpdateAddressRequest struct {
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	DetailAddress string `json:"detail_address"`
	IsDefault     bool   `json:"is_default"`
}

// UpdateAddress 更新地址
func (c *AddressController) UpdateAddress(ctx *gin.Context) {
	// 从上下文中获取用户ID
	customerID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	var req UpdateAddressRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	address, err := c.addressService.UpdateAddress(customerID.(uint), uint(id), services.UpdateAddressRequest{
		Name:          req.Name,
		Phone:         req.Phone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		DetailAddress: req.DetailAddress,
		IsDefault:     req.IsDefault,
	})
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "地址更新成功",
		"address": address,
	})
}

// DeleteAddress 删除地址
func (c *AddressController) DeleteAddress(ctx *gin.Context) {
	// 从上下文中获取用户ID
	customerID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	if err := c.addressService.DeleteAddress(customerID.(uint), uint(id)); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "地址删除成功",
	})
}

// SetDefaultAddress 设置默认地址
func (c *AddressController) SetDefaultAddress(ctx *gin.Context) {
	// 从上下文中获取用户ID
	customerID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	if err := c.addressService.SetDefaultAddress(customerID.(uint), uint(id)); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "默认地址设置成功",
	})
}

// GetDefaultAddress 获取默认地址
func (c *AddressController) GetDefaultAddress(ctx *gin.Context) {
	// 从上下文中获取用户ID
	customerID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	address, err := c.addressService.GetDefaultAddress(customerID.(uint))
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"address": address,
	})
}
