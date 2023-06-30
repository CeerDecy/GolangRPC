package pool

import "time"

// Goroutine消费者

type Worker struct {
	pool     *Pool
	task     chan func() // 任务队列
	lastTime time.Time   // 执行任务最后的时间
}

func (w *Worker) run() {
	go w.running()
}

// 执行task队列
func (w *Worker) running() {
	for t := range w.task {
		// 若t = nil 说明当前worker已经被释放掉
		if t == nil {
			return
		}
		t()
		// 执行完毕，需要将worker设置为空闲
		w.pool.PutWorker(w)
		w.pool.DecRunning()
	}
}
