package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	Text   string  `json:"text,omitempty"`
	Blocks []block `json:"blocks,omitempty"`
}

type block struct {
	Type   string      `json:"type"`
	Text   *blockText  `json:"text,omitempty"`
	Fields []blockText `json:"fields,omitempty"`
}

type blockText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (n *Notifier) Name() string {
	return n.NameValue
}

func (n *Notifier) Send(ctx context.Context, event notify.Event) error {
	body, err := json.Marshal(payload{
		Text: fmt.Sprintf("[%s] %s", event.Status, event.Summary),
		Blocks: []block{
			{
				Type: "header",
				Text: &blockText{Type: "plain_text", Text: fmt.Sprintf("[%s] %s", event.Status, event.Service)},
			},
			{
				Type: "section",
				Text: &blockText{Type: "mrkdwn", Text: event.Summary},
			},
			{
				Type: "section",
				Text: &blockText{Type: "mrkdwn", Text: formatDetails(event.Details)},
			},
			{
				Type: "context",
				Fields: []blockText{
					{Type: "mrkdwn", Text: fmt.Sprintf("*狀態*: %s", event.Status)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*時間*: %s", event.OccurredAt.Format(time.RFC3339))},
				},
			},
		},
	})
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
		return fmt.Errorf("slack status %d", resp.StatusCode)
	}
	return nil
}

func formatDetails(details string) string {
	list := format.DetailsList(details)
	if strings.TrimSpace(list) == "" {
		return "*細節*: n/a"
	}
	return "*細節*\n" + list
}
