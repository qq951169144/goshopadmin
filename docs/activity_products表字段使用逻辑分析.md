# activity_products 表字段使用逻辑分析报告

## 1. 表结构分析

**activity_products 表结构**：
| 字段名 | 数据类型 | 描述 |
| :--- | :--- | :--- |
| `id` | `int` | 主键ID |
| `activity_id` | `int` | 活动ID（外键） |
| `product_id` | `int` | 商品ID（外键） |
| `sku_id` | `int` | SKU ID（外键） |
| `merchant_id` | `int` | 商户ID（外键） |
| `product_type` | `enum` | 商品在活动中的类型：seckill-秒杀商品，redeem-兑换商品 |
| `status` | `enum` | 状态：active-激活，inactive-禁用 |
| `created_at` | `time.Time` | 创建时间 |
| `updated_at` | `time.Time` | 更新时间 |

**注**：original_price、activity_price、stock 字段已被移除，因为实际业务逻辑中未使用这些字段，价格和库存管理基于 ProductSku 表实现。

## 2. activity_price 字段使用逻辑

**实际使用场景**：
- `activity_products.price` 字段存储活动商品的**基础价格**
- 但在实际业务逻辑中，**订单创建时使用的是 `ProductSku.price` 字段**
- `ProductSku` 表中通过 `ActivityID > 0` 标识该SKU为活动商品

**相关代码**：
- `activity_order_service.go:82-83`：使用 `sku.Price` 计算商品金额
- `activity_order_service.go:89`：将 `sku.Price` 存储到订单商品项中

## 3. stock 字段使用逻辑

**实际使用场景**：
- `activity_products.stock` 字段存储活动商品的**基础库存**
- 但在实际业务逻辑中，**库存管理操作使用的是 `ProductSku.stock` 字段**

**库存管理流程**：
1. **库存检查** (`CheckActivityStock`)：检查活动商品SKU库存是否充足
2. **库存减少** (`ReduceActivityStock`)：订单创建时减少活动商品SKU库存
3. **库存恢复** (`CancelActivityOrder`)：订单取消时恢复活动商品SKU库存

**相关代码**：
- `activity_service.go:89-106`：检查活动商品SKU库存
- `activity_service.go:108-133`：减少活动商品SKU库存
- `activity_order_service.go:211-223`：取消订单时恢复活动商品SKU库存

## 4. 完整业务流程

**活动商品管理流程**：
1. **活动创建**：关联商品并设置基础价格、库存和购买限制
2. **活动商品查询**：获取活动商品列表及其SKU信息
3. **订单创建**：
   - 检查用户购买限制
   - 检查活动商品SKU库存
   - 计算订单金额（使用SKU价格）
   - 减少活动商品SKU库存
   - 创建订单及订单商品项
4. **订单取消**：
   - 恢复活动商品SKU库存
   - 更新订单状态

## 5. 总结

**activity_products 表的作用**：
- 主要作为**活动与商品的关联表**，建立活动与商品/SKU之间的关联关系
- 存储活动商品的基本配置信息，如商品类型、状态等
- 实际业务操作（价格计算、库存管理）完全基于 `ProductSku` 表

**设计意图**：
- 通过 `activity_products` 表统一管理活动与商品的关联关系
- 通过 `ProductSku` 表实现更精细的SKU级别的价格和库存管理
- 这种设计既保证了活动配置的一致性，又支持了SKU级别的灵活管理
- 移除未使用的价格和库存字段，简化表结构，提高系统维护性

## 6. 代码优化建议

1. **代码注释优化**：
   - 在相关服务方法中添加注释，明确说明价格和库存的使用逻辑
   - 特别是在库存管理相关方法中，明确说明操作的是 `ProductSku` 表的库存

2. **性能优化**：
   - 考虑为 activity_products 表添加适当的索引，提高查询性能
   - 优化活动商品关联的查询逻辑，减少数据库查询次数

通过以上分析，我们可以看到 activity_products 表在系统中主要起到配置和关联的作用，而实际的价格计算和库存管理操作则基于 ProductSku 表进行，这种设计既保证了系统的灵活性，又实现了业务逻辑的清晰分离。