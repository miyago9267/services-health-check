package check

import "context"

type Checker interface {
	Name() string
	Check(ctx context.Context) (Result, error)
}
