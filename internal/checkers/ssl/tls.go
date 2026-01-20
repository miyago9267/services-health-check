package ssl

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"services-health-check/internal/core/check"
)

type Checker struct {
	NameValue  string
	Address    string
	ServerName string
	Timeout    time.Duration
	WarnBefore time.Duration
	CritBefore time.Duration
	SkipVerify bool
}

func (c *Checker) Name() string {
	return c.NameValue
}

func (c *Checker) Check(ctx context.Context) (check.Result, error) {
	addr := c.Address
	if addr == "" {
		return check.Result{Name: c.NameValue, Status: check.StatusUnknown, Message: "缺少 address", CheckedAt: time.Now()}, fmt.Errorf("address required")
	}

	serverName := c.ServerName
	if serverName == "" {
		host, _, err := net.SplitHostPort(addr)
		if err == nil {
			serverName = host
		} else {
			serverName = addr
		}
	}

	dialer := &net.Dialer{Timeout: c.Timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: c.SkipVerify,
	})
	if err != nil {
		return check.Result{Name: c.NameValue, Status: check.StatusCrit, Message: fmt.Sprintf("TLS 連線失敗 %s (SNI %s): %v", addr, serverName, err), CheckedAt: time.Now()}, err
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return check.Result{Name: c.NameValue, Status: check.StatusUnknown, Message: "未取得憑證", CheckedAt: time.Now()}, fmt.Errorf("no peer certificates")
	}

	cert := state.PeerCertificates[0]
	until := time.Until(cert.NotAfter)

	warn := c.WarnBefore
	crit := c.CritBefore
	if warn == 0 {
		warn = 30 * 24 * time.Hour
	}
	if crit == 0 {
		crit = 7 * 24 * time.Hour
	}

	status := check.StatusOK
	message := fmt.Sprintf("憑證尚有 %s", until.Truncate(time.Hour))
	if until <= 0 {
		status = check.StatusCrit
		message = "憑證已過期"
	} else if until <= crit {
		status = check.StatusCrit
		message = fmt.Sprintf("憑證即將過期：%s", until.Truncate(time.Hour))
	} else if until <= warn {
		status = check.StatusWarn
		message = fmt.Sprintf("憑證即將過期：%s", until.Truncate(time.Hour))
	}

	return check.Result{
		Name:      c.NameValue,
		Status:    status,
		Message:   message,
		Metrics:   map[string]any{"not_after": cert.NotAfter.Format(time.RFC3339), "days_left": int(until.Hours() / 24)},
		CheckedAt: time.Now(),
	}, nil
}
