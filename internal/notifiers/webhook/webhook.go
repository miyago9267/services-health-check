package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"services-health-check/internal/core/notify"
)

type Notifier struct {
	NameValue string
	URL       string
	Timeout   time.Duration
}

func (n *Notifier) Name() string {
	return n.NameValue
}

func (n *Notifier) Send(ctx context.Context, event notify.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: n.Timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook status %d", resp.StatusCode)
	}
	return nil
}
