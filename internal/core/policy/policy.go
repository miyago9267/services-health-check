package policy

import (
	"context"

	"services-health-check/internal/core/check"
	"services-health-check/internal/core/notify"
)

type Policy interface {
	Evaluate(ctx context.Context, res check.Result) (*notify.Event, error)
}
