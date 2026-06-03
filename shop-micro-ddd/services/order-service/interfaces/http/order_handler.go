package http

import (
	"net/http"

	"order-service/domain/order"
	"order-service/interfaces/dto"

	"github.com/gin-gonic/gin"
)

// OrderHandler 订单 HTTP Handler
type OrderHandler struct {
	orderService *order.OrderService
}

// NewOrderHandler 创建订单 Handler
func NewOrderHandler(orderService *order.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// CreateOrder 创建订单
// POST /api/orders
func (h *OrderHandler) CreateOrder(ctx *gin.Context) {
	var req dto.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4001, "message": "参数错误", "data": nil})
		return
	}

	// 转换为领域层输入
	items := make([]order.ItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, order.ItemInput{
			ProductID: item.ProductID,
			SkuID:     item.SkuID,
			Quantity:  item.Quantity,
		})
	}

	orderEntity, err := h.orderService.CreateOrder(ctx.Request.Context(), req.CustomerID, req.AddressID, items, req.Remark)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 5000, "message": err.Error(), "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.toOrderResponse(orderEntity),
	})
}

// CancelOrder 取消订单
// POST /api/orders/:orderNo/cancel
func (h *OrderHandler) CancelOrder(ctx *gin.Context) {
	orderNo := ctx.Param("orderNo")
	customerID := 0 // 实际项目中从认证中间件获取
	if cid, exists := ctx.Get("customerID"); exists {
		customerID = cid.(int)
	}

	err := h.orderService.CancelOrder(ctx.Request.Context(), orderNo, customerID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 5000, "message": err.Error(), "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
}

// GetOrderDetail 获取订单详情
// GET /api/orders/:orderNo
func (h *OrderHandler) GetOrderDetail(ctx *gin.Context) {
	orderNo := ctx.Param("orderNo")
	customerID := 0 // 实际项目中从认证中间件获取
	if cid, exists := ctx.Get("customerID"); exists {
		customerID = cid.(int)
	}

	orderEntity, err := h.orderService.GetOrderDetail(ctx.Request.Context(), orderNo, customerID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 4040, "message": err.Error(), "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.toOrderResponse(orderEntity),
	})
}

// toOrderResponse 将领域实体转换为响应 DTO
func (h *OrderHandler) toOrderResponse(o *order.Order) dto.OrderResponse {
	items := make([]dto.OrderItemResponse, 0, len(o.Items()))
	for _, item := range o.Items() {
		items = append(items, dto.OrderItemResponse{
			ProductID:   item.ProductID(),
			SkuID:       item.SkuID(),
			ProductName: item.ProductName(),
			SkuAttrs:    item.SkuAttrs(),
			Price:       item.Price(),
			Quantity:    item.Quantity(),
			TotalAmount: item.TotalAmount(),
		})
	}

	var cancelledAt *string
	if o.CancelledAt() != nil {
		s := o.CancelledAt().Format("2006-01-02 15:04:05")
		cancelledAt = &s
	}

	return dto.OrderResponse{
		ID:          o.ID(),
		OrderNo:     o.OrderNo(),
		CustomerID:  o.CustomerID(),
		MerchantID:  o.MerchantID(),
		TotalAmount: o.TotalAmount(),
		Status:      string(o.Status()),
		AddressID:   o.AddressID(),
		Remark:      o.Remark(),
		Items:       items,
		CreatedAt:   o.CreatedAt().Format("2006-01-02 15:04:05"),
		CancelledAt: cancelledAt,
	}
}
