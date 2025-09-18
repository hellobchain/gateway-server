package lb

import (
	"net"
	"sync"
	"time"

	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetFileLoggerWithoutName(nil)

// HealthChecker 简易 TCP/HTTP 探活
type HealthChecker struct {
	interval time.Duration // 检测间隔
	balancer Balancer      // 负载均衡
}

func NewHealthChecker(b Balancer, interval time.Duration) *HealthChecker {
	return &HealthChecker{balancer: b, interval: interval}
}

func (h *HealthChecker) Start(inst []Instance) {
	go func() {
		for {
			healthy := h.checkAll(inst)
			h.balancer.Update(healthy)
			time.Sleep(h.interval)
			logger.Debugf("[HealthChecker] update healthy instances: %v", healthy)
		}
	}()
}

func (h *HealthChecker) checkAll(inst []Instance) []Instance {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var healthy []Instance
	logger.Debug("start health check")
	for _, v := range inst {
		wg.Add(1)
		go func(in Instance) {
			defer wg.Done()
			if h.ok(in.Addr) {
				logger.Debugf("check ok: %s", in.Addr)
				mu.Lock()
				healthy = append(healthy, in)
				mu.Unlock()
			} else {
				logger.Debugf("check failed: %s", in.Addr)
			}
		}(v)
	}
	wg.Wait()
	return healthy
}

func (h *HealthChecker) ok(addr string) bool {
	return isPortOpen(addr)
}

// isPortOpen 检测端口是否开放
func isPortOpen(addr string) bool {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
