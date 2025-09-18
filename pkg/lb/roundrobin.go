package lb

import (
	"sync/atomic"
)

type weightedRR struct {
	insts []Instance // 实例列表
	total int        // 权重总和
	pos   uint64     // 原子游标
}

func New(insts []Instance) Balancer {
	w := &weightedRR{insts: insts}
	for _, v := range insts {
		w.total += v.Weight
	}
	return w
}

func (w *weightedRR) Pick() (string, bool, bool) {
	if w.total == 0 {
		return "", false, false
	}
	pos := atomic.AddUint64(&w.pos, 1) % uint64(w.total)
	for _, v := range w.insts {
		if pos < uint64(v.Weight) {
			return getAddr(v), true, v.IsRemovePrex
		}
		pos -= uint64(v.Weight)
	}
	if len(w.insts) == 0 {
		return "", false, false
	}
	// 兜底
	defaultInstance := w.insts[0]
	return getAddr(defaultInstance), true, defaultInstance.IsRemovePrex
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
