package tracker_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/tro3373/squid-brocker/internal/config"
	"github.com/tro3373/squid-brocker/internal/tracker"
)

func mustLoadConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg, err := config.Load(filepath.Join("..", "..", "testdata", "rules_test.yaml"))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	return cfg
}

func mustNewTracker(t *testing.T) *tracker.Tracker {
	t.Helper()
	cfg := mustLoadConfig(t)
	store := tracker.NewMemoryStore()
	tr, err := tracker.New(cfg, store)
	if err != nil {
		t.Fatalf("failed to create tracker: %v", err)
	}
	return tr
}

func TestCheckAccess_FirstAccessReturnsOK(t *testing.T) {
	tr := mustNewTracker(t)
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	got := tr.CheckAccess("192.168.1.100", "www.youtube.com", now)
	if !got {
		t.Error("expected first access to be allowed")
	}
}

func TestCheckAccess_ExceedsLimitReturnsDeny(t *testing.T) {
	tr := mustNewTracker(t)
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	// 60 checks × 60s interval = 3600s = 60 minutes
	for i := range 60 {
		got := tr.CheckAccess("192.168.1.100", "www.youtube.com", now)
		if !got {
			t.Fatalf("expected access to be allowed at check %d", i)
		}
	}

	// 61st check should be denied
	got := tr.CheckAccess("192.168.1.100", "www.youtube.com", now)
	if got {
		t.Error("expected access to be denied after exceeding limit")
	}
}

func TestCheckAccess_DateRolloverResetsCounter(t *testing.T) {
	tr := mustNewTracker(t)
	day1 := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	// Use up all 60 minutes
	for range 60 {
		tr.CheckAccess("192.168.1.100", "www.youtube.com", day1)
	}
	got := tr.CheckAccess("192.168.1.100", "www.youtube.com", day1)
	if got {
		t.Fatal("expected deny on day 1")
	}

	// Next day should be allowed
	day2 := time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)
	got = tr.CheckAccess("192.168.1.100", "www.youtube.com", day2)
	if !got {
		t.Error("expected access to be allowed on new day")
	}
}

func TestCheckAccess_UnknownDeviceReturnsOK(t *testing.T) {
	tr := mustNewTracker(t)
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	got := tr.CheckAccess("192.168.1.200", "www.youtube.com", now)
	if !got {
		t.Error("expected unknown device to be allowed (fail-open)")
	}
}

func TestCheckAccess_UnknownDomainReturnsOK(t *testing.T) {
	tr := mustNewTracker(t)
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	got := tr.CheckAccess("192.168.1.100", "www.google.com", now)
	if !got {
		t.Error("expected unknown domain to be allowed")
	}
}

func TestCheckAccess_IndependentGroupTracking(t *testing.T) {
	tr := mustNewTracker(t)
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	// Use up YouTube limit (60 min)
	for range 60 {
		tr.CheckAccess("192.168.1.100", "www.youtube.com", now)
	}
	got := tr.CheckAccess("192.168.1.100", "www.youtube.com", now)
	if got {
		t.Fatal("expected YouTube to be denied")
	}

	// Social should still be allowed (30 min limit, not used yet)
	got = tr.CheckAccess("192.168.1.100", "www.tiktok.com", now)
	if !got {
		t.Error("expected social to still be allowed")
	}
}

func TestCheckAccess_DomainSuffixMatch(t *testing.T) {
	tr := mustNewTracker(t)
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	// All these should match the youtube group
	domains := []string{"www.youtube.com", "m.youtube.com", "youtube.com", "r1.googlevideo.com"}
	for _, d := range domains {
		got := tr.CheckAccess("192.168.1.100", d, now)
		if !got {
			t.Errorf("expected %q to be allowed (should match youtube group)", d)
		}
	}
}

func TestCleanOldEntries(t *testing.T) {
	tr := mustNewTracker(t)
	day1 := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	// Add some usage on day 1
	tr.CheckAccess("192.168.1.100", "www.youtube.com", day1)

	// Clean on day 2
	day2 := time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)
	tr.CleanOldEntries(day2)

	// Usage on day 2 should start fresh (full limit available)
	for i := range 60 {
		got := tr.CheckAccess("192.168.1.100", "www.youtube.com", day2)
		if !got {
			t.Fatalf("expected access to be allowed at check %d after cleanup", i)
		}
	}
}
