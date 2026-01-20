package notify

import "context"

type Notifier interface {
	Name() string
	Send(ctx context.Context, event Event) error
}
