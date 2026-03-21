package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type DomainGroup struct {
	Name    string   `yaml:"name"`
	Domains []string `yaml:"domains"`
}

type Limit struct {
	Group        string `yaml:"group"`
	DailyMinutes int    `yaml:"daily_minutes"`
}

type Rule struct {
	Device string  `yaml:"device"`
	Label  string  `yaml:"label"`
	Limits []Limit `yaml:"limits"`
}

type Config struct {
	DomainGroups         []DomainGroup `yaml:"domain_groups"`
	Rules                []Rule        `yaml:"rules"`
	CheckIntervalSeconds int           `yaml:"check_interval_seconds"`
}

func Load(path string) (*Config, error) {
	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config YAML: %w", err)
	}

	if cfg.CheckIntervalSeconds == 0 {
		cfg.CheckIntervalSeconds = 60
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	groupNames := make(map[string]bool, len(c.DomainGroups))
	for _, g := range c.DomainGroups {
		if len(g.Domains) == 0 {
			return fmt.Errorf("domain group %q has no domains", g.Name)
		}
		groupNames[g.Name] = true
	}

	for _, r := range c.Rules {
		for _, l := range r.Limits {
			if l.DailyMinutes < 0 {
				return fmt.Errorf("device %q group %q has negative daily_minutes: %d", r.Device, l.Group, l.DailyMinutes)
			}
			if !groupNames[l.Group] {
				return fmt.Errorf("device %q references unknown group %q", r.Device, l.Group)
			}
		}
	}

	return nil
}

func (c *Config) FindRule(ip string) *Rule {
	for i := range c.Rules {
		if c.Rules[i].Device == ip {
			return &c.Rules[i]
		}
	}
	return nil
}

func (c *Config) FindDomainGroup(domain string) *DomainGroup {
	for i := range c.DomainGroups {
		for _, d := range c.DomainGroups[i].Domains {
			if !strings.HasPrefix(d, ".") {
				if domain == d {
					return &c.DomainGroups[i]
				}
				continue
			}
			if strings.HasSuffix(domain, d) || domain == d[1:] {
				return &c.DomainGroups[i]
			}
		}
	}
	return nil
}
