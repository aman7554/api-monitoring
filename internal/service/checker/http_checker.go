package checker

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"pulsewatch/internal/domain"
)

type HTTPChecker struct {
	client *http.Client
}

func NewHTTPChecker() *HTTPChecker {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	return &HTTPChecker{
		client: &http.Client{
			Transport: tr,
		},
	}
}

type CheckResultDetail struct {
	Status           domain.MonitorStatus
	StatusCode       int
	LatencyMS        int
	DNSTimeMS        int
	SSLDaysRemaining *int
	ErrorMessage     string
}

func (c *HTTPChecker) Execute(ctx context.Context, m *domain.Monitor) *CheckResultDetail {
	timeout := time.Duration(m.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var dnsStart, dnsEnd time.Time
	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			dnsEnd = time.Now()
		},
	}
	reqCtx = httptrace.WithClientTrace(reqCtx, trace)

	var reqBody io.Reader
	if m.Body != "" {
		reqBody = bytes.NewBufferString(m.Body)
	}

	method := strings.ToUpper(m.Method)
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequestWithContext(reqCtx, method, m.URL, reqBody)
	if err != nil {
		return &CheckResultDetail{
			Status:       domain.MonitorStatusDown,
			ErrorMessage: fmt.Sprintf("failed to construct request: %v", err),
		}
	}

	// Parse Custom Headers
	if len(m.Headers) > 0 && string(m.Headers) != "{}" {
		var headers map[string]string
		if err := json.Unmarshal(m.Headers, &headers); err == nil {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
	}

	// Parse Auth Config
	if len(m.AuthConfig) > 0 && string(m.AuthConfig) != "{}" {
		var auth struct {
			Type     string `json:"type"` // "bearer", "basic", "header"
			Token    string `json:"token"`
			Username string `json:"username"`
			Password string `json:"password"`
			Key      string `json:"key"`
			Value    string `json:"value"`
		}
		if err := json.Unmarshal(m.AuthConfig, &auth); err == nil {
			switch strings.ToLower(auth.Type) {
			case "bearer":
				req.Header.Set("Authorization", "Bearer "+auth.Token)
			case "basic":
				req.SetBasicAuth(auth.Username, auth.Password)
			case "header":
				req.Header.Set(auth.Key, auth.Value)
			}
		}
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "PulseWatch-Monitor/1.0")
	}

	start := time.Now()
	resp, err := c.client.Do(req)
	latency := time.Since(start)

	dnsTime := 0
	if !dnsStart.IsZero() && !dnsEnd.IsZero() {
		dnsTime = int(dnsEnd.Sub(dnsStart).Milliseconds())
	}

	if err != nil {
		return &CheckResultDetail{
			Status:       domain.MonitorStatusDown,
			LatencyMS:    int(latency.Milliseconds()),
			DNSTimeMS:    dnsTime,
			ErrorMessage: fmt.Sprintf("HTTP request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	statusCode := resp.StatusCode

	// Validate expected status code
	expectedCode := m.ExpectedStatusCode
	if expectedCode == 0 {
		expectedCode = 200
	}

	if statusCode != expectedCode {
		return &CheckResultDetail{
			Status:       domain.MonitorStatusDown,
			StatusCode:   statusCode,
			LatencyMS:    int(latency.Milliseconds()),
			DNSTimeMS:    dnsTime,
			ErrorMessage: fmt.Sprintf("expected status code %d, got %d", expectedCode, statusCode),
		}
	}

	// Check response keyword if specified
	if m.ResponseKeyword != "" {
		if !strings.Contains(string(bodyBytes), m.ResponseKeyword) {
			return &CheckResultDetail{
				Status:       domain.MonitorStatusDegraded,
				StatusCode:   statusCode,
				LatencyMS:    int(latency.Milliseconds()),
				DNSTimeMS:    dnsTime,
				ErrorMessage: fmt.Sprintf("response keyword '%s' not found in body", m.ResponseKeyword),
			}
		}
	}

	return &CheckResultDetail{
		Status:     domain.MonitorStatusUp,
		StatusCode: statusCode,
		LatencyMS:  int(latency.Milliseconds()),
		DNSTimeMS:  dnsTime,
	}
}
