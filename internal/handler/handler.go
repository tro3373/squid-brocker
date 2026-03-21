package handler

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/tro3373/squid-brocker/internal/tracker"
)

// AccessChecker decides whether a device can access a domain.
type AccessChecker interface {
	CheckAccess(ip string, domain string, now time.Time) bool
}

// Run reads Squid external_acl requests from r and writes responses to w.
// Each input line: "<source_ip> <destination_domain>"
// Each output line: "OK" or "ERR"
func Run(r io.Reader, w io.Writer, checker AccessChecker) error {
	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		result := processLine(line, checker)
		if _, err := fmt.Fprintln(writer, result); err != nil {
			return fmt.Errorf("writing response: %w", err)
		}
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("flushing response: %w", err)
		}
	}

	return scanner.Err()
}

func processLine(line string, checker AccessChecker) string {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "ERR"
	}

	ip := parts[0]
	domain := parts[1]

	if checker.CheckAccess(ip, domain, time.Now()) {
		return "OK"
	}
	return "ERR"
}

// Ensure *tracker.Tracker satisfies AccessChecker at compile time.
var _ AccessChecker = (*tracker.Tracker)(nil)
