package lb

import (
	"sync/atomic"
)

type weightedRR struct {
	insts []Instance
	total int    // 权重总和
	pos   uint64 // 原子游标
}

func New(insts []Instance) Balancer {
	w := &weightedRR{insts: insts}
	for _, v := range insts {
		w.total += v.Weight
	}
	return w
}

func (w *weightedRR) Pick() (string, bool) {
	if w.total == 0 {
		return "", false
	}
	pos := atomic.AddUint64(&w.pos, 1) % uint64(w.total)
	for _, v := range w.insts {
		if pos < uint64(v.Weight) {
			return getAddr(v), true
		}
		pos -= uint64(v.Weight)
	}
	// 兜底
	return getAddr(w.insts[0]), true
}

func getAddr(inst Instance) string {
	if inst.Protocol == "" {
		inst.Protocol = "http"
	}
	return inst.Protocol + "://" + inst.Addr
}

func (w *weightedRR) Update(insts []Instance) {
	w.insts = insts
	w.total = 0
	for _, v := range insts {
		w.total += v.Weight
	}
}
