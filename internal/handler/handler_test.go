package handler_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/tro3373/squid-brocker/internal/handler"
)

type mockChecker struct {
	allowAll bool
	calls    []checkCall
}

type checkCall struct {
	IP     string
	Domain string
}

func (m *mockChecker) CheckAccess(ip string, domain string, _ time.Time) bool {
	m.calls = append(m.calls, checkCall{IP: ip, Domain: domain})
	return m.allowAll
}

func TestRun_AllowedAccess(t *testing.T) {
	input := "192.168.1.100 www.youtube.com\n"
	r := strings.NewReader(input)
	var w bytes.Buffer
	checker := &mockChecker{allowAll: true}

	err := handler.Run(r, &w, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(w.String())
	if got != "OK" {
		t.Errorf("expected OK, got %q", got)
	}
	if len(checker.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(checker.calls))
	}
	if checker.calls[0].IP != "192.168.1.100" {
		t.Errorf("expected IP 192.168.1.100, got %q", checker.calls[0].IP)
	}
	if checker.calls[0].Domain != "www.youtube.com" {
		t.Errorf("expected domain www.youtube.com, got %q", checker.calls[0].Domain)
	}
}

func TestRun_DeniedAccess(t *testing.T) {
	input := "192.168.1.100 www.youtube.com\n"
	r := strings.NewReader(input)
	var w bytes.Buffer
	checker := &mockChecker{allowAll: false}

	err := handler.Run(r, &w, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(w.String())
	if got != "ERR" {
		t.Errorf("expected ERR, got %q", got)
	}
}

func TestRun_MultipleLines(t *testing.T) {
	input := "192.168.1.100 www.youtube.com\n192.168.1.101 www.tiktok.com\n"
	r := strings.NewReader(input)
	var w bytes.Buffer
	checker := &mockChecker{allowAll: true}

	err := handler.Run(r, &w, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(w.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(lines), lines)
	}
	for i, line := range lines {
		if line != "OK" {
			t.Errorf("line %d: expected OK, got %q", i, line)
		}
	}
	if len(checker.calls) != 2 {
		t.Errorf("expected 2 calls, got %d", len(checker.calls))
	}
}

func TestRun_MalformedInput(t *testing.T) {
	input := "only-one-field\n"
	r := strings.NewReader(input)
	var w bytes.Buffer
	checker := &mockChecker{allowAll: true}

	err := handler.Run(r, &w, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(w.String())
	if got != "ERR" {
		t.Errorf("expected ERR for malformed input, got %q", got)
	}
	if len(checker.calls) != 0 {
		t.Errorf("expected no checker calls for malformed input, got %d", len(checker.calls))
	}
}

func TestRun_EmptyLines(t *testing.T) {
	input := "\n\n192.168.1.100 www.youtube.com\n\n"
	r := strings.NewReader(input)
	var w bytes.Buffer
	checker := &mockChecker{allowAll: true}

	err := handler.Run(r, &w, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(w.String())
	if got != "OK" {
		t.Errorf("expected single OK, got %q", got)
	}
}

func TestRun_EmptyInput(t *testing.T) {
	r := strings.NewReader("")
	var w bytes.Buffer
	checker := &mockChecker{allowAll: true}

	err := handler.Run(r, &w, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Len() != 0 {
		t.Errorf("expected no output, got %q", w.String())
	}
}
