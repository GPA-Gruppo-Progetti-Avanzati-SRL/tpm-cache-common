package cachelks

import "context"

type LinkedService interface {
	Name() string
	Type() string
	Set(ctx context.Context, db int, key string, value interface{}) error
	Get(ctx context.Context, db int, key string) (interface{}, error)
}
