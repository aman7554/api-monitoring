package checker

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"pulsewatch/internal/domain"
)

type DNSChecker struct{}

func NewDNSChecker() *DNSChecker {
	return &DNSChecker{}
}

func (c *DNSChecker) Execute(ctx context.Context, m *domain.Monitor) *CheckResultDetail {
	host := m.URL
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		if parsed, err := url.Parse(host); err == nil {
			host = parsed.Hostname()
		}
	}

	timeout := time.Duration(m.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resolver := &net.Resolver{}
	start := time.Now()
	addrs, err := resolver.LookupHost(reqCtx, host)
	latency := time.Since(start)

	if err != nil {
		return &CheckResultDetail{
			Status:       domain.MonitorStatusDown,
			LatencyMS:    int(latency.Milliseconds()),
			DNSTimeMS:    int(latency.Milliseconds()),
			ErrorMessage: fmt.Sprintf("DNS lookup failed for %s: %v", host, err),
		}
	}

	if len(addrs) == 0 {
		return &CheckResultDetail{
			Status:       domain.MonitorStatusDown,
			LatencyMS:    int(latency.Milliseconds()),
			DNSTimeMS:    int(latency.Milliseconds()),
			ErrorMessage: fmt.Sprintf("DNS resolved no IP addresses for %s", host),
		}
	}

	return &CheckResultDetail{
		Status:     domain.MonitorStatusUp,
		StatusCode: 200,
		LatencyMS:  int(latency.Milliseconds()),
		DNSTimeMS:  int(latency.Milliseconds()),
	}
}
