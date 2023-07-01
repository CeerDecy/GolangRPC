package pool

import (
	"fmt"
	"time"
)

// Goroutine消费者

type Worker struct {
	pool     *Pool
	task     chan func() // 任务队列
	lastTime time.Time   // 执行任务最后的时间
}

func (w *Worker) run() {
	w.pool.IncRunning()
	go w.running()
}

// 执行task队列
func (w *Worker) running() {
	defer func() {
		w.pool.DecRunning()
		w.pool.workerCache.Put(w)
		if err := recover(); err != nil {
			if w.pool.PanicHandler != nil {
				w.pool.PanicHandler()
			} else {
				fmt.Println(err)
			}
			w.pool.cond.Signal()
		}
	}()
	for t := range w.task {
		// 若t = nil 说明当前worker已经被释放掉
		if t == nil {
			w.pool.workerCache.Put(w)
			return
		}
		t()
		// 执行完毕，需要将worker设置为空闲
		w.pool.PutWorker(w)
	}
}
