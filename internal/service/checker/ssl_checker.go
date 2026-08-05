package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"pulsewatch/internal/domain"
)

type SSLChecker struct{}

func NewSSLChecker() *SSLChecker {
	return &SSLChecker{}
}

func (c *SSLChecker) Execute(ctx context.Context, m *domain.Monitor) *CheckResultDetail {
	targetURL := m.URL
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return &CheckResultDetail{
			Status:       domain.MonitorStatusDown,
			ErrorMessage: fmt.Sprintf("invalid URL for SSL check: %v", err),
		}
	}

	host := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		port = "443"
	}
	address := fmt.Sprintf("%s:%s", host, port)

	timeout := time.Duration(m.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	dialer := &net.Dialer{Timeout: timeout}
	start := time.Now()
	conn, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	})
	latency := time.Since(start)

	if err != nil {
		return &CheckResultDetail{
			Status:       domain.MonitorStatusDown,
			LatencyMS:    int(latency.Milliseconds()),
			ErrorMessage: fmt.Sprintf("TLS connection failed: %v", err),
		}
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return &CheckResultDetail{
			Status:       domain.MonitorStatusDown,
			LatencyMS:    int(latency.Milliseconds()),
			ErrorMessage: "no SSL certificates presented",
		}
	}

	leafCert := certs[0]
	now := time.Now()
	daysRemaining := int(leafCert.NotAfter.Sub(now).Hours() / 24)

	status := domain.MonitorStatusUp
	var errMsg string

	if now.After(leafCert.NotAfter) {
		status = domain.MonitorStatusDown
		errMsg = fmt.Sprintf("SSL certificate expired on %s", leafCert.NotAfter.Format(time.RFC3339))
	} else if daysRemaining < 14 {
		status = domain.MonitorStatusDegraded
		errMsg = fmt.Sprintf("SSL certificate expiring soon (%d days remaining)", daysRemaining)
	}

	return &CheckResultDetail{
		Status:           status,
		StatusCode:       200,
		LatencyMS:        int(latency.Milliseconds()),
		SSLDaysRemaining: &daysRemaining,
		ErrorMessage:     errMsg,
	}
}
