package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"services-health-check/internal/core/check"

	"github.com/likexian/whois"
	"github.com/likexian/whois-parser"
)

type ExpiryChecker struct {
	NameValue    string
	Domain       string
	Timeout      time.Duration
	WarnBefore   time.Duration
	CritBefore   time.Duration
	RDAPBaseURL  string
	RDAPBaseURLs []string
}

func (c *ExpiryChecker) Name() string {
	return c.NameValue
}

func (c *ExpiryChecker) Check(ctx context.Context) (check.Result, error) {
	if strings.TrimSpace(c.Domain) == "" {
		return check.Result{Name: c.NameValue, Status: check.StatusUnknown, Message: "缺少 domain", CheckedAt: time.Now()}, fmt.Errorf("domain required")
	}

	cctx, cancel := c.withJitter(ctx)
	defer cancel()

	exp, err := c.lookupExpiration(cctx)
	if err != nil {
		return check.Result{Name: c.NameValue, Status: check.StatusCrit, Message: fmt.Sprintf("查詢失敗（%s）: %s", c.Domain, err.Error()), CheckedAt: time.Now()}, err
	}
	until := time.Until(exp)
	remaining := formatDurationDHMS(until)

	warn := c.WarnBefore
	crit := c.CritBefore
	if warn == 0 {
		warn = 30 * 24 * time.Hour
	}
	if crit == 0 {
		crit = 7 * 24 * time.Hour
	}

	status := check.StatusOK
	message := fmt.Sprintf("網域尚有 %s（%s）", remaining, c.Domain)
	if until <= 0 {
		status = check.StatusCrit
		message = fmt.Sprintf("網域已過期（%s）", c.Domain)
	} else if until <= crit {
		status = check.StatusCrit
		message = fmt.Sprintf("網域即將過期（%s）：%s", c.Domain, remaining)
	} else if until <= warn {
		status = check.StatusWarn
		message = fmt.Sprintf("網域即將過期（%s）：%s", c.Domain, remaining)
	}

	return check.Result{
		Name:      c.NameValue,
		Status:    status,
		Message:   message,
		Metrics:   map[string]any{"expiration": exp.Format(time.RFC3339)},
		CheckedAt: time.Now(),
	}, nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (c *ExpiryChecker) withJitter(ctx context.Context) (context.Context, context.CancelFunc) {
	jitter := time.Duration(rand.Intn(10)) * time.Second
	timer := time.NewTimer(jitter)
	select {
	case <-timer.C:
	case <-ctx.Done():
		timer.Stop()
	}
	return context.WithCancel(ctx)
}

func (c *ExpiryChecker) lookupExpiration(ctx context.Context) (time.Time, error) {
	if exp, err := c.lookupWhois(); err == nil {
		return exp, nil
	}
	return c.lookupRDAP(ctx)
}

func (c *ExpiryChecker) lookupWhois() (time.Time, error) {
	client := whois.NewClient()
	if c.Timeout > 0 {
		client.SetTimeout(c.Timeout)
	}

	raw, err := client.Whois(c.Domain)
	if err != nil {
		return time.Time{}, err
	}

	info, err := whoisparser.Parse(raw)
	if err != nil {
		return time.Time{}, err
	}
	if info.Domain == nil || info.Domain.ExpirationDateInTime == nil {
		return time.Time{}, fmt.Errorf("whois 無法取得到期日")
	}
	return *info.Domain.ExpirationDateInTime, nil
}

func (c *ExpiryChecker) lookupRDAP(ctx context.Context) (time.Time, error) {
	bases := c.rdapBases()
	var lastErr error
	for _, base := range bases {
		exp, err := c.lookupRDAPBase(ctx, base)
		if err == nil {
			return exp, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return time.Time{}, lastErr
	}
	return time.Time{}, fmt.Errorf("rdap 無可用來源")
}

func (c *ExpiryChecker) rdapBases() []string {
	var bases []string
	if strings.TrimSpace(c.RDAPBaseURL) != "" {
		bases = append(bases, c.RDAPBaseURL)
	}
	for _, b := range c.RDAPBaseURLs {
		if strings.TrimSpace(b) == "" {
			continue
		}
		bases = append(bases, b)
	}
	if len(bases) == 0 {
		bases = append(bases, "https://rdap.org")
	}
	return bases
}

func (c *ExpiryChecker) lookupRDAPBase(ctx context.Context, base string) (time.Time, error) {
	client := &http.Client{Timeout: c.Timeout}
	if client.Timeout == 0 {
		client.Timeout = 10 * time.Second
	}
	base = strings.TrimRight(base, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/domain/"+c.Domain, nil)
	if err != nil {
		return time.Time{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return time.Time{}, fmt.Errorf("rdap HTTP %d", resp.StatusCode)
	}

	var payload struct {
		Events []struct {
			Action string `json:"eventAction"`
			Date   string `json:"eventDate"`
		} `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return time.Time{}, err
	}
	for _, ev := range payload.Events {
		if strings.EqualFold(ev.Action, "expiration") && ev.Date != "" {
			t, err := time.Parse(time.RFC3339, ev.Date)
			if err != nil {
				return time.Time{}, err
			}
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("rdap 無到期日")
}

func formatDurationDHMS(d time.Duration) string {
	seconds := int64(d.Seconds())
	if seconds < 0 {
		seconds = -seconds
	}
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%dd%dh%dm%ds", days, hours, minutes, secs)
}
