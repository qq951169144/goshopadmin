# 待修bug和优化列表

## 1. 日志文件分割问题

- **问题描述**：日志没有按照固定大小分割文件，程序启动时异常切割文件，客户端访问后疑似正常
- **状态**：未检查出原因
- **优先级**：中

## 2. MQ连接协程优化

- **问题描述**：协程开启在大并发下可能会暴涨
- **修改方案**：参考以下文件
  - `d:\code\goshopadmin\.trae\documents\goroutine_analysis_plan.md`
  - `d:\code\goshopadmin\.trae\documents\goroutine_optimization_flow.md`
  - `d:\code\goshopadmin\.trae\documents\goroutine_optimization_implementation.md`
- **优先级**：高

## 4. WebSocket引入

- **问题描述**：需要引入WebSocket功能
- **参考方案**：`d:\code\goshopadmin\.trae\documents\websocket-notification-plan.md`
- **优先级**：最低

## 5. 多商户的订单处理方案

- **问题描述**：需要实现多商户的订单处理，目前默认为1了，
- **优先级**：低
- **理由**：该项目着重关注redis，mq，websocket，nginx抗压架构上

<br />

## 7. 检查sku新增/更新库存的时候有没有更新缓存

- **问题描述**：检查sku新增/更新库存的时候有没有更新缓存
- **优先级**：低

## 8. 删除后台活动(新建和修改)页面设置关联sku物品的时候，不能在活动处修改sku的价格和库存，只能回到商品管理处理

- **问题描述**：删除后台活动设置关联sku物品的时候，不能在活动处修改sku的价格和库存，只能回到商品管理处理
- **优先级**：低

