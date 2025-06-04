package scheduler

type AdaptiveLoadBalancer struct {
	roundRobin int
}

func NewAdaptiveLoadBalancer() *AdaptiveLoadBalancer {
	return &AdaptiveLoadBalancer{}
}

func (lb *AdaptiveLoadBalancer) NextIndex(total int) int {
	if total <= 0 {
		return 0
	}
	idx := lb.roundRobin % total
	lb.roundRobin++
	return idx
}
