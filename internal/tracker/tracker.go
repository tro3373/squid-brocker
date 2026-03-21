package tracker

import (
	"fmt"
	"sync"
	"time"

	"github.com/tro3373/squid-brocker/internal/config"
)

// Tracker tracks cumulative access time per device per domain group.
type Tracker struct {
	cfg   *config.Config
	store Store
	mu    sync.Mutex
	state State
}

func New(cfg *config.Config, store Store) (*Tracker, error) {
	state, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("loading state: %w", err)
	}
	return &Tracker{
		cfg:   cfg,
		store: store,
		state: state,
	}, nil
}

// CheckAccess determines if a device is allowed to access a domain.
// Returns true if access is allowed.
func (t *Tracker) CheckAccess(ip string, domain string, now time.Time) bool {
	rule := t.cfg.FindRule(ip)
	if rule == nil {
		return true
	}

	group := t.cfg.FindDomainGroup(domain)
	if group == nil {
		return true
	}

	var limit *config.Limit
	for i := range rule.Limits {
		if rule.Limits[i].Group == group.Name {
			limit = &rule.Limits[i]
			break
		}
	}
	if limit == nil {
		return true
	}

	today := now.Format("2006-01-02")
	key := UsageKey{
		Device: ip,
		Group:  group.Name,
		Date:   today,
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	current := t.state[key]
	if current.Label == "" {
		current.Label = rule.Label
	}
	limitSeconds := limit.DailyMinutes * 60

	if current.Seconds >= limitSeconds {
		if t.state[key].Label == "" {
			t.state[key] = current
			_ = t.store.Save(t.state)
		}
		return false
	}

	t.state[key] = UsageValue{
		Seconds: current.Seconds + t.cfg.CheckIntervalSeconds,
		Label:   current.Label,
	}
	_ = t.store.Save(t.state)

	return true
}

// CleanOldEntries removes state entries from dates before today.
func (t *Tracker) CleanOldEntries(now time.Time) {
	today := now.Format("2006-01-02")

	t.mu.Lock()
	defer t.mu.Unlock()

	for k := range t.state {
		if k.Date != today {
			delete(t.state, k)
		}
	}
	_ = t.store.Save(t.state)
}
