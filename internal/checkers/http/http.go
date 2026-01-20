package httpcheck

import (
	"context"
	"net/http"
	"time"

	"services-health-check/internal/core/check"
)

type Checker struct {
	NameValue string
	URL       string
	Timeout   time.Duration
}

func (c *Checker) Name() string {
	return c.NameValue
}

func (c *Checker) Check(ctx context.Context) (check.Result, error) {
	client := &http.Client{Timeout: c.Timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return check.Result{Name: c.NameValue, Status: check.StatusUnknown, Message: "請求建立失敗: " + err.Error(), CheckedAt: time.Now()}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return check.Result{Name: c.NameValue, Status: check.StatusCrit, Message: "連線失敗: " + err.Error(), CheckedAt: time.Now()}, err
	}
	defer resp.Body.Close()

	status := check.StatusOK
	if resp.StatusCode >= 400 {
		status = check.StatusCrit
	}

	return check.Result{
		Name:      c.NameValue,
		Status:    status,
		Message:   "HTTP 狀態: " + resp.Status,
		Metrics:   map[string]any{"status_code": resp.StatusCode},
		CheckedAt: time.Now(),
	}, nil
}
