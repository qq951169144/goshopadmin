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
- **完成状态**：TODO:首次使用superpowers-zh完成该功能，待功能测试

## 4. WebSocket引入

- **问题描述**：需要引入WebSocket功能
- **参考方案**：`d:\code\goshopadmin\.trae\documents\websocket-notification-plan.md`
- **优先级**：最低

## 5. 多商户的订单处理方案

- **问题描述**：需要实现多商户的订单处理，目前默认为1了，
- **优先级**：低
- **理由**：该项目着重关注redis，mq，websocket，nginx抗压架构上
- **参考方案**：`D:\code\goshopadmin\docs\多商户订单处理方案.md`

<br />

## 6. 检查sku新增/更新库存的时候有没有更新缓存

- **问题描述**：检查sku新增/更新库存的时候有没有更新缓存
- **优先级**：低

##

## 7. 日志文件记录长记录问题

- **问题描述**：日志对于一些大文件请求的写入非常消耗资源，我希望过滤掉这些日志的写入，或者忽略掉那些请求和返回是图片，文件等大文件的response,request不记录
- **优先级**：高
- **完成状态**：TODO:使用superpowers-zh完成该功能，待功能测试


## 8. 引入更详细monitor.go监控

- **问题描述**：引入更详细monitor.go监控
- **优先级**：高
- **完成状态**：TODO:使用superpowers-zh完成该功能，待功能测试


## 9. 引入elk

- **问题描述**：引入elk
- **优先级**：高
- **完成状态**：TODO:使用superpowers-zh完成该功能，待功能测试


## 10. 订单管理点击进去，没有任何请求search-service的接口

- **问题描述**：订单管理点击进去，没有任何请求search-service的接口
- **优先级**：高
- **完成状态**：待测试
- **文件列表**：
  -`D:\code\goshopadmin\frontend\src\views\Home.vue`
  -`D:\code\goshopadmin\search-service\controllers\order_controller.go`
  -`D:\code\goshopadmin\search-service`


## 11. api/search/customer/orders接口有问题

- **问题描述**：api/search/customer/orders接口有问题，带了authorization值验证，经过查看和其他接口带的token是一致的，auth.go返回4012无效token
- **优先级**：高
- **完成状态**：待测试
- **文件列表**：
  -`D:\code\goshopadmin\search-service\middleware\auth.go`


## 12. 综合搜素侧边栏删除

- **问题描述**：综合搜素侧边栏删除,包含他跳转的页面
- **优先级**：高
- **完成状态**：待测试
- **文件列表**：
  -`D:\code\goshopadmin\frontend\src\views\Home.vue`

## 13. AdminAuth 和 CustomerAuth 要从Authorization header提取token

- **问题描述**：AdminAuth 和 CustomerAuth 要从Authorization header提取token，验证方法要和backend和shop-backend的相同
- **优先级**：高
- **完成状态**：待测试
- **文件列表**：
  -`D:\code\goshopadmin\search-service\middleware\auth.go`
  -`D:\code\goshopadmin\backend\middleware\auth.go`
  -`D:\code\goshopadmin\shop-backend\middleware\auth.go`