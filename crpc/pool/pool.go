package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type sig struct{}

const DefaultExpire = 3

var ErrorInvalidCap = errors.New("pool cap can not <= 0")
var ErrorInvalidExpire = errors.New("pool expire can not <= 0")

// Pool 协程池
type Pool struct {
	cap     int32         // 最大Worker容量
	running int32         // 正在运行的Worker数量
	workers []*Worker     // 空闲的Worker
	expire  time.Duration // 过期时间
	release chan sig      // 释放信号，释放后pool就不能使用了
	lock    sync.Mutex    // 保证pool内部资源的并发安全
	once    sync.Once     // 释放资源只能调用一次，不能多次调用
}

func NewTimePool(cap int32, expire int) (*Pool, error) {
	if cap <= 0 {
		return nil, ErrorInvalidCap
	}
	if expire <= 0 {
		return nil, ErrorInvalidExpire
	}
	return &Pool{
		cap:     cap,
		expire:  time.Duration(expire) * time.Second,
		release: make(chan sig, 1),
	}, nil
}

func NewPool(cap int32) (*Pool, error) {
	return NewTimePool(cap, DefaultExpire)
}

func (p *Pool) Submit(task func()) error {
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
		worker := &Worker{pool: p, task: make(chan func(), 1), lastTime: time.Now()}
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
