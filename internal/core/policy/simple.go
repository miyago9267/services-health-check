package policy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"services-health-check/internal/core/check"
	"services-health-check/internal/core/notify"
)

type SimplePolicy struct {
	Cooldown         time.Duration
	NotifyOnRecovery bool

	mu           sync.Mutex
	lastStatus   map[string]check.Status
	lastNotified map[string]time.Time
}

func NewSimplePolicy(cooldown time.Duration, notifyOnRecovery bool) *SimplePolicy {
	return &SimplePolicy{
		Cooldown:         cooldown,
		NotifyOnRecovery: notifyOnRecovery,
		lastStatus:       make(map[string]check.Status),
		lastNotified:     make(map[string]time.Time),
	}
}

func (p *SimplePolicy) Evaluate(ctx context.Context, res check.Result) (*notify.Event, error) {
	_ = ctx
	p.mu.Lock()
	defer p.mu.Unlock()

	p.lastStatus[res.Name] = res.Status

	now := time.Now()
	if p.Cooldown > 0 {
		if lastNotifyAt, ok := p.lastNotified[res.Name]; ok {
			if now.Sub(lastNotifyAt) < p.Cooldown {
				return nil, nil
			}
		}
	}

	p.lastNotified[res.Name] = now
	summary := fmt.Sprintf("%s 狀態：%s", res.Name, res.Status)
	if res.Status == check.StatusUnknown {
		summary = fmt.Sprintf("%s 掃描失敗", res.Name)
	}
	return &notify.Event{
		Service:    res.Name,
		Status:     string(res.Status),
		Summary:    summary,
		Details:    res.Message,
		Labels:     map[string]string{"status": string(res.Status)},
		OccurredAt: now,
	}, nil
}
