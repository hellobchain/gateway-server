package lb

// Instance 后端节点
type Instance struct {
	Addr         string
	Protocol     string
	Weight       int
	IsRemovePrex bool
}

// Balancer 负载均衡器
type Balancer interface {
	Pick() (string, bool, bool) // 返回选中的 addr
	Update([]Instance)          // 批量更新节点
}
