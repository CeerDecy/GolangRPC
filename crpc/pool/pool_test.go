package pool

import (
	"math"
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	_   = 1 << (10 * iota)
	KiB = 1 << (10 * iota)
	MiB = 1 << (10 * iota)
)

const (
	Param    = 100
	PoolSize = 1000
	TestSize = 10000
	n        = 1000000
)

var curMem uint64

const (
	RunTimes          = 1000000
	BenchParam        = 10
	DefaultExpireTime = 10 * time.Second
)

func demoFunc() {
	time.Sleep(time.Duration(BenchParam) * time.Millisecond)
}

func TestNoPool(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			demoFunc()
			wg.Done()
		}()
	}
	wg.Wait()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("no pool memory used : %d MB", curMem)
}
func TestHasPool(t *testing.T) {
	pool, _ := NewPool(math.MaxInt32)
	defer pool.Release()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = pool.Submit(func() {
			defer wg.Done()
			demoFunc()
		})
	}
	wg.Wait()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("has pool memory used : %d MB", curMem)
}
