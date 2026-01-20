package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"services-health-check/internal/core/check"
)

type TokenChecker struct {
	NameValue string
	Token     string
	Timeout   time.Duration
	BaseURL   string
}

type tokenVerifyResponse struct {
	Success bool           `json:"success"`
	Errors  []errorMessage `json:"errors"`
	Result  struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"result"`
}

type errorMessage struct {
	Message string `json:"message"`
}

func (c *TokenChecker) Name() string {
	return c.NameValue
}

func (c *TokenChecker) Check(ctx context.Context) (check.Result, error) {
	if strings.TrimSpace(c.Token) == "" {
		return check.Result{Name: c.NameValue, Status: check.StatusUnknown, Message: "缺少 Cloudflare token", CheckedAt: time.Now()}, fmt.Errorf("token required")
	}

	timeout := c.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	client := &http.Client{Timeout: timeout}
	url := "https://api.cloudflare.com/client/v4/user/tokens/verify"
	if strings.TrimSpace(c.BaseURL) != "" {
		url = strings.TrimRight(c.BaseURL, "/") + "/user/tokens/verify"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return check.Result{Name: c.NameValue, Status: check.StatusCrit, Message: "建立請求失敗: " + err.Error(), CheckedAt: time.Now()}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := client.Do(req)
	if err != nil {
		return check.Result{Name: c.NameValue, Status: check.StatusCrit, Message: "連線失敗: " + err.Error(), CheckedAt: time.Now()}, err
	}
	defer resp.Body.Close()

	var payload tokenVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return check.Result{Name: c.NameValue, Status: check.StatusCrit, Message: "解析回應失敗: " + err.Error(), CheckedAt: time.Now()}, err
	}

	status := check.StatusOK
	message := "token 正常"
	if !payload.Success {
		status = check.StatusCrit
		message = "token 驗證失敗: " + joinErrors(payload.Errors)
	} else if strings.ToLower(payload.Result.Status) != "active" && payload.Result.Status != "" {
		status = check.StatusWarn
		message = fmt.Sprintf("token 狀態: %s", payload.Result.Status)
	}

	return check.Result{
		Name:      c.NameValue,
		Status:    status,
		Message:   message,
		Metrics:   map[string]any{"status": payload.Result.Status, "id": payload.Result.ID},
		CheckedAt: time.Now(),
	}, nil
}

func joinErrors(errs []errorMessage) string {
	if len(errs) == 0 {
		return "token verify failed"
	}
	var out []string
	for _, e := range errs {
		if strings.TrimSpace(e.Message) != "" {
			out = append(out, e.Message)
		}
	}
	if len(out) == 0 {
		return "token verify failed"
	}
	return strings.Join(out, "; ")
}
