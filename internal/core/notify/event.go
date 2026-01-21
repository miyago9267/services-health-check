package notify

import "time"

type Event struct {
	Service    string
	Type       string
	Status     string
	Summary    string
	Details    string
	Labels     map[string]string
	OccurredAt time.Time
}
