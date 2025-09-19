package breaker

import (
	"errors"
	"sync/atomic"
	"time"
)

type State int32

const (
	StateClosed   State = 0
	StateOpen     State = 1
	StateHalfOpen State = 2
)

var ErrBreakerOpen = errors.New("circuit breaker open")

type Settings struct {
	Enabled               bool          // 熔断器是否开启
	MaxRequests           uint32        // 半开时最大探测请求数
	Interval              time.Duration // 统计窗口
	Timeout               time.Duration // 熔断后多久进入半开
	ErrorPercentThreshold float64       // 错误率阈值（0-100）
	MinRequestAmount      uint32        // 最小请求数才触发错误率计算
}

func NewDefaultSettings() Settings {
	return Settings{
		Enabled:               true,
		MaxRequests:           5,
		Interval:              10 * time.Second,
		Timeout:               5 * time.Second,
		ErrorPercentThreshold: 50,
		MinRequestAmount:      10,
	}
}

type SreBreaker struct {
	settings Settings
	state    int32 // atomic

	// 滑动窗口计数器
	lastResetTime int64 // unix nano
	reqAll        uint32
	reqFail       uint32
}

func New(settings Settings) *SreBreaker {
	b := &SreBreaker{settings: settings}
	return b
}

func (b *SreBreaker) Enabled() bool {
	return b.settings.Enabled
}

// 对外唯一入口
func (b *SreBreaker) Do(fn func() error) error {
	for {
		st := State(atomic.LoadInt32(&b.state))
		switch st {
		case StateClosed:
			return b.call(fn)
		case StateOpen:
			return ErrBreakerOpen
		case StateHalfOpen:
			if atomic.LoadUint32(&b.reqAll) >= b.settings.MaxRequests {
				return ErrBreakerOpen
			}
			err := b.call(fn)
			if err == nil {
				b.onSuccess()
			} else {
				b.onFail()
			}
			return err
		}
	}
}

/* ---------- 内部 ---------- */
func (b *SreBreaker) call(fn func() error) error {
	b.onRequest()
	err := fn()
	if err != nil {
		b.onFail()
		return err
	}
	b.onSuccess()
	return nil
}

func (b *SreBreaker) onRequest() {
	b.lazyReset()
	atomic.AddUint32(&b.reqAll, 1)
}

func (b *SreBreaker) onFail() {
	atomic.AddUint32(&b.reqFail, 1)
	b.evaluate()
}

func (b *SreBreaker) onSuccess() {
	b.evaluate()
}

func (b *SreBreaker) evaluate() {
	all := atomic.LoadUint32(&b.reqAll)
	fail := atomic.LoadUint32(&b.reqFail)
	if all < b.settings.MinRequestAmount {
		return
	}
	percent := float64(fail) / float64(all) * 100
	if percent >= b.settings.ErrorPercentThreshold {
		atomic.StoreInt32(&b.state, int32(StateOpen))
		// 启动倒计时：Timeout 后进入半开
		time.AfterFunc(b.settings.Timeout, func() {
			atomic.StoreInt32(&b.state, int32(StateHalfOpen))
			atomic.StoreUint32(&b.reqAll, 0)
			atomic.StoreUint32(&b.reqFail, 0)
		})
	}
}

// 滑动窗口重置
func (b *SreBreaker) lazyReset() {
	now := time.Now().UnixNano()
	if now-atomic.LoadInt64(&b.lastResetTime) > int64(b.settings.Interval) {
		atomic.StoreInt64(&b.lastResetTime, now)
		atomic.StoreUint32(&b.reqAll, 0)
		atomic.StoreUint32(&b.reqFail, 0)
		atomic.StoreInt32(&b.state, int32(StateClosed))
	}
}
