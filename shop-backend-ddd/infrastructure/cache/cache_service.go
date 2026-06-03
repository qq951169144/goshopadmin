package cache

// CacheService 缓存服务（简化版，演示接口抽象）
type CacheService struct {
	// 实际项目中这里会有 Redis 客户端
}

// NewCacheService 创建缓存服务
func NewCacheService() *CacheService {
	return &CacheService{}
}

// DeleteOrderCache 删除订单缓存
func (s *CacheService) DeleteOrderCache(orderNo string) error {
	// 实际实现：删除 Redis 中的订单缓存
	return nil
}
