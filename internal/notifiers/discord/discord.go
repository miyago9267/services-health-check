package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"services-health-check/internal/core/notify"
	"services-health-check/internal/notifiers/format"
)

type Notifier struct {
	NameValue string
	URL       string
	Username  string
	Timeout   time.Duration
}

type payload struct {
	Content  string  `json:"content,omitempty"`
	Username string  `json:"username,omitempty"`
	Embeds   []embed `json:"embeds,omitempty"`
}

type embed struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Color       int          `json:"color,omitempty"`
	Fields      []embedField `json:"fields,omitempty"`
	Timestamp   string       `json:"timestamp,omitempty"`
}

type embedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func (n *Notifier) Name() string {
	return n.NameValue
}

func (n *Notifier) Send(ctx context.Context, event notify.Event) error {
	body, err := json.Marshal(payload{
		Embeds: []embed{
			{
				Title:       fmt.Sprintf("[%s] %s", event.Status, event.Service),
				Description: event.Summary,
				Color:       statusColor(event.Status),
				Fields: []embedField{
					{Name: "Details", Value: formatDetails(event.Details), Inline: false},
				},
				Timestamp: event.OccurredAt.Format(time.RFC3339),
			},
		},
		Username: n.Username,
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func statusColor(status string) int {
	switch strings.ToUpper(status) {
	case "OK":
		return 0x2ECC71
	case "WARN":
		return 0xF1C40F
	case "CRIT":
		return 0xE74C3C
	default:
		return 0x95A5A6
	}
}

func formatDetails(details string) string {
	normalized := format.DetailsList(details)
	if len(normalized) > 900 {
		normalized = normalized[:900] + "\n- ...（已截斷）"
	}
	return "```\n" + normalized + "\n```"
}
