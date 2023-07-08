package breaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

// Counts 计数器
type Counts struct {
	Requests            uint64
	TotalSuccess        uint64
	TotalFailures       uint64
	ConsecutiveSuccess  uint64
	ConsecutiveFailures uint64
}

// OnRequest 有一个请求进来则+1
func (c *Counts) OnRequest() {
	c.Requests++
}

// OnSuccess 成功累计，失败归零
func (c *Counts) OnSuccess() {
	c.TotalSuccess++
	c.ConsecutiveSuccess++
	c.ConsecutiveFailures = 0
}

// OnFailures 失败累计，成功归零
func (c *Counts) OnFailures() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccess = 0
}

// Clear 清空计数器
func (c *Counts) Clear() {
	c.Requests = 0
	c.TotalFailures = 0
	c.TotalSuccess = 0
	c.ConsecutiveFailures = 0
	c.ConsecutiveSuccess = 0
}

// Settings 断路器配置表
type Settings struct {
	Name          string                                  // 名称
	MaxRequest    uint64                                  // 最大请求数量
	Interval      time.Duration                           // 间隔时间
	Timeout       time.Duration                           // 超时时间
	ReadyToTrip   func(counts Counts) bool                // 执行熔断
	OnStateChange func(name string, from State, to State) //状态变更
	IsSuccess     func(err error) bool                    // 是否成功
	FallBack      func(err error) (any, error)
}

// CircuitBreaker 断路器
type CircuitBreaker struct {
	name          string
	maxRequests   uint64
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(counts Counts) bool
	isSuccessful  func(err error) bool
	onStateChange func(name string, from State, to State)
	mutex         sync.Mutex
	state         State
	generation    uint64
	counts        Counts
	expiry        time.Time
	fallBack      func(err error) (any, error)
}

// NewGeneration 重置断路器
func (b *CircuitBreaker) NewGeneration() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.generation++
	b.counts.Clear()
	var zero time.Time
	switch b.state {
	case StateClosed:
		if b.interval == 0 {
			b.expiry = zero
		} else {
			b.expiry = time.Now().Add(b.interval)
		}
	case StateHalfOpen:
		b.expiry = time.Now()
	case StateOpen:
		b.expiry = time.Now().Add(b.timeout)
	}
}

func (b *CircuitBreaker) Execute(req func() (any, error)) (any, error) {
	// 请求之前判断是否执行断路器
	generation, err := b.beforeRequest()
	if err != nil {
		if b.fallBack != nil {
			return b.fallBack(err)
		}
		return nil, err
	}

	res, err := req()

	b.counts.OnRequest()
	// 请求之后判断，当前状态是否需要变更
	b.afterRequest(generation, b.isSuccessful(err))
	return res, err
}

func (b *CircuitBreaker) beforeRequest() (uint64, error) {
	state, _ := b.currentState(time.Now())
	if state == StateOpen {
		return 0, errors.New("breaker has been opened")
	}
	if state == StateHalfOpen {
		if b.counts.Requests >= b.maxRequests {
			return 0, errors.New("too many requests")
		}
	}

	return b.generation, nil
}

func (b *CircuitBreaker) afterRequest(before uint64, success bool) {
	state, _ := b.currentState(time.Now())
	if b.generation != before {
		return
	}
	if success {
		b.OnSuccess(state)
	} else {
		b.OnFailures(state)
	}
}

func (b *CircuitBreaker) currentState(now time.Time) (State, uint64) {
	switch b.state {
	case StateClosed:
		if !b.expiry.IsZero() && b.expiry.Before(now) {
			b.NewGeneration()
		}
	case StateOpen:
		if b.expiry.Before(now) {
			b.setState(StateHalfOpen)
		}
	}
	return b.state, b.generation
}

func (b *CircuitBreaker) setState(state State) {
	if b.state == state {
		return
	}
	before := b.state
	b.state = state
	b.NewGeneration()
	if b.onStateChange != nil {
		b.onStateChange(b.name, before, state)
	}

}

func (b *CircuitBreaker) OnSuccess(state State) {
	switch state {
	case StateClosed:
		b.counts.OnSuccess()
		if b.readyToTrip(b.counts) {
			b.setState(StateOpen)
		}
	case StateHalfOpen:
		b.counts.OnSuccess()
		if b.readyToTrip(b.counts) {
			b.setState(StateOpen)
		}
	case StateOpen:
		b.counts.OnSuccess()
	}
}

func (b *CircuitBreaker) OnFailures(state State) {
	switch state {
	case StateClosed:
		b.counts.OnFailures()
		if b.counts.ConsecutiveFailures >= b.maxRequests {
			b.setState(StateOpen)
		}
	case StateHalfOpen:
		b.counts.OnFailures()
		if b.counts.ConsecutiveFailures >= b.maxRequests {
			b.setState(StateOpen)
		}
	}
}

func NewCircuitBreaker(settings Settings) *CircuitBreaker {
	breaker := &CircuitBreaker{
		name:          settings.Name,
		interval:      settings.Interval,
		onStateChange: settings.OnStateChange,
		fallBack:      settings.FallBack,
	}
	if settings.MaxRequest == 0 {
		breaker.maxRequests = 1
	} else {
		breaker.maxRequests = settings.MaxRequest
	}
	if settings.Timeout == 0 {
		breaker.timeout = 20 * time.Second
	} else {
		breaker.timeout = settings.Timeout
	}
	if settings.ReadyToTrip == nil {
		breaker.readyToTrip = func(counts Counts) bool {
			return counts.ConsecutiveFailures > 5
		}
	} else {
		breaker.readyToTrip = settings.ReadyToTrip
	}
	if settings.IsSuccess == nil {
		breaker.isSuccessful = func(err error) bool {
			return err == nil
		}
	} else {
		breaker.isSuccessful = settings.IsSuccess
	}
	breaker.NewGeneration()
	return breaker
}
