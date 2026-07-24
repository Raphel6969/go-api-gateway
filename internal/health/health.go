package health

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Raphel6969/api-gateway/internal/loadbalancer"
)

type HealthChecker struct {
	allTargets []string
	lb         loadbalancer.LoadBalancer
	interval   time.Duration
	client     *http.Client
	log        *slog.Logger
}

func NewHealthCheck(targets []string, lb loadbalancer.LoadBalancer, interval time.Duration, log *slog.Logger) *HealthChecker {
	return &HealthChecker{
		allTargets: targets,
		lb:         lb,
		interval:   interval,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		log: log,
	}
}

func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.interval)
	go func() {
		for range ticker.C {
			hc.checkAll()
		}
	}()
}

func (hc *HealthChecker) checkAll() {
	var healthy []string

	for _, target := range hc.allTargets {
		resp, err := hc.client.Get(target + "/")
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode <= 400 {
			healthy = append(healthy, target)
			_ = resp.Body.Close()
		} else {
			if resp != nil {
				_ = resp.Body.Close()
			}
			hc.log.Warn("Backend servicae is UNHEALTHY", "target", target, "error", err)
		}
	}

	hc.lb.UpdateTargets(healthy)
}
