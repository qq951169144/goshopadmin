package utils

import (
	"sync"
	"time"
)

// Task 任务结构，包含要执行的函数和超时时间
type Task struct {
	fn      func()        // 要执行的任务函数
	timeout time.Duration // 任务超时时间，0表示不限制
}

// WorkerPool 工作协程池
type WorkerPool struct {
	tasks          chan Task        // 任务队列
	workers        int              // 当前工作协程数量
	minWorkers     int              // 最小工作协程数
	maxWorkers     int              // 最大工作协程数
	mu             sync.RWMutex     // 互斥锁，保护共享状态
	wg             sync.WaitGroup   // 等待组，用于优雅关闭
	quit           chan struct{}    // 退出信号通道
	stats          PoolStats        // 统计信息
	scaleCheckInt  time.Duration    // 伸缩检查间隔
}

// PoolStats 工作池统计信息
type PoolStats struct {
	TasksSubmitted  int64 // 已提交任务总数
	TasksCompleted  int64 // 已完成任务总数
	TasksFailed     int64 // 失败任务总数
	CurrentQueueLen int   // 当前队列长度
	CurrentWorkers  int   // 当前工作协程数
}

// NewWorkerPool 创建工作池
func NewWorkerPool(minWorkers, maxWorkers, queueSize int) *WorkerPool {
	if minWorkers <= 0 {
		minWorkers = 2
	}
	if maxWorkers <= minWorkers {
		maxWorkers = minWorkers * 2
	}
	if queueSize <= 0 {
		queueSize = 1000
	}

	pool := &WorkerPool{
		tasks:         make(chan Task, queueSize),
		minWorkers:    minWorkers,
		maxWorkers:    maxWorkers,
		quit:          make(chan struct{}),
		scaleCheckInt: 10 * time.Second,
	}

	// 启动初始工作协程
	for i := 0; i < minWorkers; i++ {
		pool.startWorker()
	}

	// 启动动态伸缩协程
	go pool.scaleLoop()

	return pool
}

// startWorker 启动一个工作协程
func (p *WorkerPool) startWorker() {
	p.mu.Lock()
	if p.workers >= p.maxWorkers {
		p.mu.Unlock()
		return
	}
	p.workers++
	p.mu.Unlock()

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		for {
			select {
			case task := <-p.tasks:
				p.executeTask(task)
			case <-p.quit:
				p.mu.Lock()
				p.workers--
				p.mu.Unlock()
				return
			}
		}
	}()
}

// executeTask 执行任务
func (p *WorkerPool) executeTask(task Task) {
	defer func() {
		if r := recover(); r != nil {
			p.mu.Lock()
			p.stats.TasksFailed++
			p.mu.Unlock()
			Error("任务执行 panic: %v", r)
		}
	}()

	if task.timeout > 0 {
		done := make(chan struct{})
		go func() {
			task.fn()
			close(done)
		}()

		select {
		case <-done:
			p.mu.Lock()
			p.stats.TasksCompleted++
			p.mu.Unlock()
		case <-time.After(task.timeout):
			p.mu.Lock()
			p.stats.TasksFailed++
			p.mu.Unlock()
			Warn("任务执行超时")
		}
	} else {
		task.fn()
		p.mu.Lock()
		p.stats.TasksCompleted++
		p.mu.Unlock()
	}
}

// Submit 提交一个任务到工作池
func (p *WorkerPool) Submit(fn func()) {
	p.tasks <- Task{fn: fn, timeout: 0}
	p.mu.Lock()
	p.stats.TasksSubmitted++
	p.mu.Unlock()
}

// SubmitWithTimeout 提交一个带超时的任务到工作池
func (p *WorkerPool) SubmitWithTimeout(fn func(), timeout time.Duration) {
	p.tasks <- Task{fn: fn, timeout: timeout}
	p.mu.Lock()
	p.stats.TasksSubmitted++
	p.mu.Unlock()
}

// Close 优雅关闭工作池
func (p *WorkerPool) Close() {
	close(p.quit)
	p.wg.Wait()
}

// GetStats 获取工作池统计信息
func (p *WorkerPool) GetStats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return PoolStats{
		TasksSubmitted:  p.stats.TasksSubmitted,
		TasksCompleted:  p.stats.TasksCompleted,
		TasksFailed:     p.stats.TasksFailed,
		CurrentQueueLen: len(p.tasks),
		CurrentWorkers:  p.workers,
	}
}

// scaleLoop 动态调整工作协程数量
func (p *WorkerPool) scaleLoop() {
	ticker := time.NewTicker(p.scaleCheckInt)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.scaleWorkers()
		case <-p.quit:
			return
		}
	}
}

// scaleWorkers 根据队列负载动态调整工作协程数量
func (p *WorkerPool) scaleWorkers() {
	p.mu.Lock()
	defer p.mu.Unlock()

	queueLen := len(p.tasks)
	workers := p.workers

	// 如果队列长度超过工作协程数的2倍，增加工作协程
	if queueLen > workers*2 && workers < p.maxWorkers {
		toAdd := min(queueLen/workers, p.maxWorkers-workers)
		for i := 0; i < toAdd; i++ {
			go p.startWorker()
		}
		Info("工作池扩容: %d -> %d", workers, workers+toAdd)
	}

	// 如果队列为空且工作协程数超过最小值，减少工作协程
	if queueLen == 0 && workers > p.minWorkers {
		toRemove := min(workers-p.minWorkers, workers/2)
		for i := 0; i < toRemove; i++ {
			p.tasks <- Task{fn: func() {}, timeout: 0}
		}
		Info("工作池缩容: %d -> %d", workers, workers-toRemove)
	}
}

// min 返回较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
