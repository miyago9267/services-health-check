package check

import "time"

type Status string

const (
	StatusOK      Status = "OK"
	StatusWarn    Status = "WARN"
	StatusCrit    Status = "CRIT"
	StatusUnknown Status = "UNKNOWN"
)

type Result struct {
	Name      string
	Status    Status
	Message   string
	Metrics   map[string]any
	CheckedAt time.Time
}
