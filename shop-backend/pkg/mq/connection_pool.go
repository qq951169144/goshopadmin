package mq

import (
	"errors"
	"sync"
	"time"

	"shop-backend/utils"
)

// PooledConnection 包装后的连接，包含元数据
type PooledConnection struct {
	conn     *Connection // 底层MQ连接
	lastUsed time.Time   // 最后使用时间
	useCount int         // 使用次数计数
	inUse    bool        // 是否正在使用中
}

// ConnectionPool MQ连接池
type ConnectionPool struct {
	connections    chan *PooledConnection // 空闲连接通道
	mu             sync.RWMutex           // 互斥锁，保护共享状态
	minConns       int                    // 最小连接数
	maxConns       int                    // 最大连接数
	idleTimeout    time.Duration          // 连接空闲超时时间
	healthCheckInt time.Duration          // 健康检查间隔
	maxUseCount    int                    // 连接最大使用次数
	closed         bool                   // 连接池是否已关闭
	stats          PoolStats              // 统计信息
	quit           chan struct{}          // 退出信号通道
}

// PoolStats 连接池统计信息
type PoolStats struct {
	TotalCreated  int64 // 总创建连接数
	TotalReleased int64 // 总释放连接数
	CurrentActive int64 // 当前活跃连接数
	CurrentIdle   int64 // 当前空闲连接数
	TotalErrors   int64 // 总错误数
}

// NewConnectionPool 创建MQ连接池
func NewConnectionPool(minConns, maxConns int) (*ConnectionPool, error) {
	if minConns <= 0 {
		minConns = 5
	}
	if maxConns <= minConns {
		maxConns = minConns * 10
	}

	pool := &ConnectionPool{
		connections:    make(chan *PooledConnection, maxConns),
		minConns:       minConns,
		maxConns:       maxConns,
		idleTimeout:    5 * time.Minute,
		healthCheckInt: 30 * time.Second,
		maxUseCount:    1000,
		closed:         false,
		quit:           make(chan struct{}),
	}

	// 预创建最小数量的连接
	for i := 0; i < minConns; i++ {
		conn, err := NewConnection()
		if err != nil {
			utils.Error("预创建MQ连接失败: %v", err)
			continue
		}
		pool.connections <- &PooledConnection{
			conn:     conn,
			lastUsed: time.Now(),
			useCount: 0,
			inUse:    false,
		}
		pool.stats.TotalCreated++
	}

	// 启动健康检查协程
	go pool.healthCheck()

	return pool, nil
}

// Get 从连接池获取一个可用连接
func (p *ConnectionPool) Get() (*Connection, error) {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return nil, errors.New("connection pool closed")
	}
	p.mu.RUnlock()

	select {
	case pc := <-p.connections:
		// 检查连接是否过期或需要替换
		if pc.useCount >= p.maxUseCount || time.Since(pc.lastUsed) > p.idleTimeout {
			pc.conn.Close()
			conn, err := NewConnection()
			if err != nil {
				p.stats.TotalErrors++
				return nil, err
			}
			pc = &PooledConnection{
				conn:     conn,
				lastUsed: time.Now(),
				useCount: 0,
				inUse:    false,
			}
			p.stats.TotalCreated++
		}
		pc.inUse = true
		pc.lastUsed = time.Now()
		p.stats.CurrentActive++
		p.stats.CurrentIdle--
		return pc.conn, nil
	default:
		// 没有空闲连接，检查是否可以创建新连接
		p.mu.Lock()
		if int(p.stats.TotalCreated-p.stats.TotalReleased) < p.maxConns {
			conn, err := NewConnection()
			if err != nil {
				p.mu.Unlock()
				p.stats.TotalErrors++
				return nil, err
			}
			p.stats.TotalCreated++
			p.stats.CurrentActive++
			p.mu.Unlock()
			return conn, nil
		}
		p.mu.Unlock()
		return nil, errors.New("connection pool exhausted")
	}
}

// Put 将连接归还到连接池
func (p *ConnectionPool) Put(conn *Connection) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed || conn == nil {
		if conn != nil {
			conn.Close()
		}
		return
	}

	pc := &PooledConnection{
		conn:     conn,
		lastUsed: time.Now(),
		useCount: 0,
		inUse:    false,
	}

	select {
	case p.connections <- pc:
		p.stats.CurrentActive--
		p.stats.CurrentIdle++
	default:
		conn.Close()
		p.stats.TotalReleased++
	}
}

// Close 关闭连接池，释放所有资源
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	close(p.quit)
	p.mu.Unlock()

	close(p.connections)
	for pc := range p.connections {
		pc.conn.Close()
		p.stats.TotalReleased++
	}
}

// GetStats 获取连接池统计信息
func (p *ConnectionPool) GetStats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats
}

// healthCheck 健康检查协程
func (p *ConnectionPool) healthCheck() {
	ticker := time.NewTicker(p.healthCheckInt)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.performHealthCheck()
		case <-p.quit:
			return
		}
	}
}

// performHealthCheck 执行健康检查
func (p *ConnectionPool) performHealthCheck() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	// 确保有足够的空闲连接
	for len(p.connections) < p.minConns {
		conn, err := NewConnection()
		if err != nil {
			utils.Error("健康检查创建连接失败: %v", err)
			p.stats.TotalErrors++
			break
		}
		p.connections <- &PooledConnection{
			conn:     conn,
			lastUsed: time.Now(),
			useCount: 0,
			inUse:    false,
		}
		p.stats.TotalCreated++
		p.stats.CurrentIdle++
	}
}
