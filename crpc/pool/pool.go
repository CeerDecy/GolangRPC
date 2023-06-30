package pool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type sig struct{}

const DefaultExpire = 2

var ErrorInvalidCap = errors.New("pool cap can not <= 0")
var ErrorInvalidExpire = errors.New("pool expire can not <= 0")
var ErrorPoolReleased = errors.New("pool has been released")

// Pool 协程池
type Pool struct {
	cap         int32         // 最大Worker容量
	running     int32         // 正在运行的Worker数量
	workers     []*Worker     // 空闲的Worker
	expire      time.Duration // 过期时间
	release     chan sig      // 释放信号，释放后pool就不能使用了
	lock        sync.Mutex    // 保证pool内部资源的并发安全
	once        sync.Once     // 释放资源只能调用一次，不能多次调用
	workerCache sync.Pool     // 缓存
}

func NewTimePool(cap int32, expire int) (*Pool, error) {
	if cap <= 0 {
		return nil, ErrorInvalidCap
	}
	if expire <= 0 {
		return nil, ErrorInvalidExpire
	}
	p := &Pool{
		cap:     cap,
		expire:  time.Duration(expire) * time.Second,
		release: make(chan sig, 1),
	}
	p.workerCache.New = func() any {
		return &Worker{
			pool: p,
			task: make(chan func(), 1),
		}
	}
	go p.expireWorker()
	return p, nil
}

func NewPool(cap int32) (*Pool, error) {
	return NewTimePool(cap, DefaultExpire)
}

func (p *Pool) Submit(task func()) error {
	if len(p.release) > 0 {
		return ErrorPoolReleased
	}
	w := p.GetWorker()
	w.task <- task
	p.IncRunning()
	return nil
}

// GetWorker 获取Worker
func (p *Pool) GetWorker() *Worker {
	// 1. 若有空闲的Worker直接获取
	// 2. 如果没有空闲的Worker，则新建一个Worker
	// 3. 若正在运行的Workers 大于 Pool的容量，则阻塞等待worker释放
	n := len(p.workers) - 1
	if n >= 0 {
		p.lock.Lock()
		defer p.lock.Unlock()
		worker := p.workers[n]
		p.workers = p.workers[:n]
		return worker
	}
	// 还没达到容量，可以新建一个Worker
	if p.running < p.cap {
		c := p.workerCache.Get()
		var worker *Worker
		if c == nil {
			worker = &Worker{
				pool: p,
				task: make(chan func(), 1),
			}
		} else {
			worker = c.(*Worker)
		}
		worker.run()
		return worker
	}
	for {
		p.lock.Lock()
		n = len(p.workers) - 1
		if n < 0 {
			p.lock.Unlock()
			continue
		}
		worker := p.workers[n]
		p.workers = p.workers[:n]
		p.lock.Unlock()
		return worker
	}
}

func (p *Pool) IncRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) PutWorker(w *Worker) {
	w.lastTime = time.Now()
	p.lock.Lock()
	defer p.lock.Unlock()
	p.workers = append(p.workers, w)
}

func (p *Pool) DecRunning() {
	atomic.AddInt32(&p.running, -1)
}

// Release 释放协程池子
func (p *Pool) Release() {
	p.once.Do(func() {
		p.lock.Lock()
		for i := range p.workers {
			p.workers[i].task = nil
			p.workers[i].pool = nil
			p.workers[i] = nil
		}
		p.workers = nil
		p.lock.Unlock()
		p.release <- sig{}
	})
}

// IsClosed 是否已经关闭
func (p *Pool) IsClosed() bool {
	return len(p.release) > 0
}

// Restart Pool释放过后可以重启
func (p *Pool) Restart() bool {
	if !p.IsClosed() {
		return true
	}
	_ = <-p.release
	p.expireWorker()
	return true
}

// 定期清理空闲的Worker
func (p *Pool) expireWorker() {
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		if p.IsClosed() {
			break
		}
		p.lock.Lock()
		// 遍历空闲Worker，判断其最后运行时间的差值是否大于expire,若大于则将其删除
		n := len(p.workers) - 1
		if n >= 0 {
			for i := range p.workers {
				if time.Now().Sub(p.workers[i].lastTime) >= p.expire {
					n = i
					p.workers[i].task <- nil
				} else {
					break
				}
			}
			if n >= len(p.workers)-1 {
				p.workers = p.workers[:0]
			} else {
				p.workers = p.workers[n+1:]
			}
			fmt.Printf("清除完成，running:%d,workers:%v\n", p.running, p.workers)
		}
		p.lock.Unlock()
	}
}
