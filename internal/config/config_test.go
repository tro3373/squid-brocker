package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tro3373/squid-brocker/internal/config"
)

func TestLoad_ValidConfig(t *testing.T) {
	cfg, err := config.Load(filepath.Join("..", "..", "testdata", "rules_test.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.DomainGroups) != 2 {
		t.Errorf("expected 2 domain groups, got %d", len(cfg.DomainGroups))
	}
	if cfg.DomainGroups[0].Name != "youtube" {
		t.Errorf("expected first group name 'youtube', got %q", cfg.DomainGroups[0].Name)
	}
	if len(cfg.DomainGroups[0].Domains) != 3 {
		t.Errorf("expected 3 domains in youtube group, got %d", len(cfg.DomainGroups[0].Domains))
	}

	if len(cfg.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(cfg.Rules))
	}
	if cfg.Rules[0].Device != "192.168.1.100" {
		t.Errorf("expected device '192.168.1.100', got %q", cfg.Rules[0].Device)
	}
	if cfg.Rules[0].Label != "kids-tablet" {
		t.Errorf("expected label 'kids-tablet', got %q", cfg.Rules[0].Label)
	}
	if len(cfg.Rules[0].Limits) != 2 {
		t.Errorf("expected 2 limits for first rule, got %d", len(cfg.Rules[0].Limits))
	}
	if cfg.Rules[0].Limits[0].Group != "youtube" {
		t.Errorf("expected group 'youtube', got %q", cfg.Rules[0].Limits[0].Group)
	}
	if cfg.Rules[0].Limits[0].DailyMinutes != 60 {
		t.Errorf("expected 60 daily minutes, got %d", cfg.Rules[0].Limits[0].DailyMinutes)
	}

	if cfg.CheckIntervalSeconds != 60 {
		t.Errorf("expected check_interval_seconds 60, got %d", cfg.CheckIntervalSeconds)
	}
}

func TestLoad_DefaultCheckInterval(t *testing.T) {
	content := []byte(`
domain_groups:
  - name: test
    domains:
      - .example.com
rules: []
`)
	tmp := t.TempDir()
	path := filepath.Join(tmp, "rules.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckIntervalSeconds != 60 {
		t.Errorf("expected default check_interval_seconds 60, got %d", cfg.CheckIntervalSeconds)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.yaml")
	if err := os.WriteFile(path, []byte(":::invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoad_EmptyDomainGroup(t *testing.T) {
	content := []byte(`
domain_groups:
  - name: empty
    domains: []
rules: []
`)
	tmp := t.TempDir()
	path := filepath.Join(tmp, "rules.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for empty domain group")
	}
}

func TestLoad_NegativeDailyMinutes(t *testing.T) {
	content := []byte(`
domain_groups:
  - name: test
    domains:
      - .example.com
rules:
  - device: 192.168.1.1
    limits:
      - group: test
        daily_minutes: -10
`)
	tmp := t.TempDir()
	path := filepath.Join(tmp, "rules.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for negative daily_minutes")
	}
}

func TestLoad_UnknownGroupReference(t *testing.T) {
	content := []byte(`
domain_groups:
  - name: test
    domains:
      - .example.com
rules:
  - device: 192.168.1.1
    limits:
      - group: nonexistent
        daily_minutes: 60
`)
	tmp := t.TempDir()
	path := filepath.Join(tmp, "rules.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for unknown group reference")
	}
}

func TestConfig_FindRule(t *testing.T) {
	cfg, err := config.Load(filepath.Join("..", "..", "testdata", "rules_test.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := cfg.FindRule("192.168.1.100")
	if rule == nil {
		t.Fatal("expected to find rule for 192.168.1.100")
	}
	if rule.Device != "192.168.1.100" {
		t.Errorf("expected device '192.168.1.100', got %q", rule.Device)
	}

	rule = cfg.FindRule("192.168.1.200")
	if rule != nil {
		t.Error("expected nil for unknown device")
	}
}

func TestConfig_FindDomainGroup(t *testing.T) {
	cfg, err := config.Load(filepath.Join("..", "..", "testdata", "rules_test.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	group := cfg.FindDomainGroup("www.youtube.com")
	if group == nil {
		t.Fatal("expected to find group for www.youtube.com")
	}
	if group.Name != "youtube" {
		t.Errorf("expected group 'youtube', got %q", group.Name)
	}

	group = cfg.FindDomainGroup("m.youtube.com")
	if group == nil {
		t.Fatal("expected to find group for m.youtube.com")
	}

	group = cfg.FindDomainGroup("youtube.com")
	if group == nil {
		t.Fatal("expected to find group for youtube.com (exact suffix match)")
	}

	group = cfg.FindDomainGroup("www.google.com")
	if group != nil {
		t.Error("expected nil for unmatched domain")
	}
}
