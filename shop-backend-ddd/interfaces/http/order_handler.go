package http

import (
	"net/http"
	"strconv"

	"shop-backend-ddd/domain/order"
	"shop-backend-ddd/interfaces/dto"

	"github.com/gin-gonic/gin"
)

// OrderHandler 订单 HTTP Handler
// 只负责 HTTP 协议相关的事情：绑定参数、调用服务、返回响应
// 不包含任何业务逻辑
type OrderHandler struct {
	orderService *order.OrderService
}

// NewOrderHandler 创建订单 Handler（构造器注入领域服务）
func NewOrderHandler(orderService *order.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// CreateOrder 创建订单
// POST /api/orders
func (h *OrderHandler) CreateOrder(ctx *gin.Context) {
	var req dto.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, dto.APIResponse{
			Code:    4001,
			Message: "参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 从上下文获取客户ID（由认证中间件设置）
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		ctx.JSON(http.StatusOK, dto.APIResponse{
			Code:    4010,
			Message: "未认证",
			Data:    nil,
		})
		return
	}

	// 转换 DTO → 领域服务输入
	items := make([]order.ItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, order.ItemInput{
			ProductID: item.ProductID,
			SkuID:     item.SkuID,
			Quantity:  item.Quantity,
		})
	}

	// 调用领域服务
	orderEntity, err := h.orderService.CreateOrder(ctx.Request.Context(), customerID.(int), req.AddressID, items, req.Remark)
	if err != nil {
		ctx.JSON(http.StatusOK, dto.APIResponse{
			Code:    5000,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// 领域模型 → 响应 DTO
	resp := dto.CreateOrderResponse{
		OrderNo:     orderEntity.OrderNo(),
		TotalAmount: orderEntity.TotalAmount().StringFixed(2),
		Status:      string(orderEntity.Status()),
		PaymentURL:  "/api/payment/fake-pay?order_id=" + orderEntity.OrderNo(),
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Code:    0,
		Message: "success",
		Data:    resp,
	})
}

// GetOrderDetail 获取订单详情
// GET /api/orders/:orderNo
func (h *OrderHandler) GetOrderDetail(ctx *gin.Context) {
	orderNo := ctx.Param("orderNo")
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		ctx.JSON(http.StatusOK, dto.APIResponse{
			Code:    4010,
			Message: "未认证",
			Data:    nil,
		})
		return
	}

	orderEntity, err := h.orderService.GetOrderDetail(ctx.Request.Context(), orderNo, customerID.(int))
	if err != nil {
		ctx.JSON(http.StatusOK, dto.APIResponse{
			Code:    4044,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// 领域模型 → 响应 DTO
	items := make([]dto.OrderItemResponse, 0, len(orderEntity.Items()))
	for _, item := range orderEntity.Items() {
		items = append(items, dto.OrderItemResponse{
			ProductID:   item.ProductID(),
			ProductName: item.ProductName(),
			SkuID:       item.SkuID(),
			Price:       item.Price().StringFixed(2),
			Quantity:    item.Quantity(),
			TotalAmount: item.TotalAmount().StringFixed(2),
		})
	}

	resp := dto.OrderDetailResponse{
		OrderID:     strconv.Itoa(orderEntity.ID()),
		OrderNo:     orderEntity.OrderNo(),
		TotalAmount: orderEntity.TotalAmount().StringFixed(2),
		Status:      string(orderEntity.Status()),
		CreatedAt:   orderEntity.CreatedAt().Format("2006-01-02 15:04:05"),
		Items:       items,
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Code:    0,
		Message: "success",
		Data:    resp,
	})
}

// CancelOrder 取消订单
// PUT /api/orders/:orderNo/cancel
func (h *OrderHandler) CancelOrder(ctx *gin.Context) {
	orderNo := ctx.Param("orderNo")
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		ctx.JSON(http.StatusOK, dto.APIResponse{
			Code:    4010,
			Message: "未认证",
			Data:    nil,
		})
		return
	}

	err := h.orderService.CancelOrder(ctx.Request.Context(), orderNo, customerID.(int))
	if err != nil {
		ctx.JSON(http.StatusOK, dto.APIResponse{
			Code:    5000,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Code:    0,
		Message: "订单已取消",
		Data:    nil,
	})
}
