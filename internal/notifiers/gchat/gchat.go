package gchat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"services-health-check/internal/core/notify"
	"services-health-check/internal/notifiers/format"
)

type Notifier struct {
	NameValue string
	URL       string
	Timeout   time.Duration
}

type payload struct {
	Text string `json:"text"`
}

func (n *Notifier) Name() string {
	return n.NameValue
}

func (n *Notifier) Send(ctx context.Context, event notify.Event) error {
	text := fmt.Sprintf("[%s] %s\n%s", event.Status, event.Summary, format.DetailsList(event.Details))
	body, err := json.Marshal(payload{Text: text})
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: n.Timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, bytes.NewReader(body))
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
		return fmt.Errorf("gchat status %d", resp.StatusCode)
	}
	return nil
}
