package loadbalancer

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrNoTargets = errors.New("no healthy targets available")

type LoadBalancer interface {
	Next() (string, error)
	UpdateTargets(targets []string)
}

type RoundRobin struct {
	targets []string
	index   uint64
	mu      sync.RWMutex
}

func NewRoundRobin(targets []string) *RoundRobin {
	return &RoundRobin{
		targets: targets,
	}
}

func (rr *RoundRobin) Next() (string, error) {
	rr.mu.RLock()
	defer rr.mu.RUnlock()
	if len(rr.targets) == 0 {
		return "", ErrNoTargets
	}

	n := atomic.AddUint64(&rr.index, 1)
	target := rr.targets[(n-1)%uint64(len(rr.targets))]

	return target, nil
}

func (rr *RoundRobin) UpdateTargets(newTargets []string) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.targets = newTargets
}
