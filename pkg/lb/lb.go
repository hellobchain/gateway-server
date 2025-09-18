package lb

// Instance 后端节点
type Instance struct {
	Addr         string // 后端地址
	Protocol     string // 协议
	Weight       int    // 权重
	IsRemovePrex bool   // 是否移除前缀
}

// Balancer 负载均衡器
type Balancer interface {
	Pick() (string, bool, bool) // 返回选中的 addr
	Update([]Instance)          // 批量更新节点
}
