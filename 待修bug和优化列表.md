# 待修bug和优化列表

## 1. 日志文件分割问题

- **问题描述**：日志没有按照固定大小分割文件，程序启动时异常切割文件，客户端访问后疑似正常
- **状态**：未检查出原因
- **优先级**：中

## 2. MQ连接协程优化

- **问题描述**：协程开启在大并发下可能会暴涨
- **修改方案**：参考以下文件
  - `D:\code\goshopadmin\.trae\documents\goroutine_optimization_enhanced_plan.md`
  - `D:\code\goshopadmin\.trae\documents\goroutine_optimization_implementation_plan.md`
- **优先级**：高
- **完成状态**：首次使用superpowers-zh完成该功能，待功能测试

## 4. WebSocket引入

- **问题描述**：需要引入WebSocket功能
- **参考方案**：`d:\code\goshopadmin\.trae\documents\websocket-notification-plan.md`
- **优先级**：最低

## 5. 多商户的订单处理方案

- **问题描述**：需要实现多商户的订单处理，目前默认为1了，
- **优先级**：低
- **理由**：该项目着重关注redis，mq，websocket，nginx抗压架构上

<br />

## 6. 检查sku新增/更新库存的时候有没有更新缓存

- **问题描述**：检查sku新增/更新库存的时候有没有更新缓存
- **优先级**：低

##

## 7. 日志文件记录长记录问题

- **问题描述**：日志对于一些大文件请求的写入非常消耗资源，我希望过滤掉这些日志的写入，或者忽略掉那些请求和返回是图片，文件等大文件的response,request不记录
- **优先级**：高
- **完成状态**：使用superpowers-zh完成该功能，待功能测试


## 8. 引入更详细monitor.go监控

- **问题描述**：引入更详细monitor.go监控
- **优先级**：高
- **完成状态**：使用superpowers-zh完成该功能，待功能测试


## 8. 后台改变总库存的显示，显示sku的库存的总和

- **问题描述**：后台改变总库存的显示，显示product_skus的库存的总和
- **优先级**：高
- **涉及文件**：D:\code\goshopadmin\backend\services\product_service.go
- **完成状态**：已完成 - 在GetProductsByMerchantID中添加Preload("Skus")，计算所有启用SKU的库存总和并赋值给product.Stock

## 9. 删除ProductSKUs.vue中活动专用SKU，活动选择器两个

- **问题描述**：删除ProductSKUs.vue中活动专用SKU，活动选择器两个
- **优先级**：高
- **涉及文件**：D:\code\goshopadmin\frontend\src\views\products\ProductSKUs.vue
- **完成状态**：已完成 - 删除了表格中的"活动专用"列、表单中的活动专用SKU复选框和关联活动选择器，以及相关的数据字段和方法