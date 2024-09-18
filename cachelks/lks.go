package cachelks

import (
	"context"
)

type LinkedService interface {
	Name() string
	Type() string
	Set(ctx context.Context, key string, value interface{}, opts ...CacheOption) error
	Get(ctx context.Context, key string, opts ...CacheOption) (interface{}, error)
}

type CacheOptions struct {
	Namespace string
}

type CacheOption func(*CacheOptions)

func WithNamespace(namespace string) CacheOption {
	return func(o *CacheOptions) {
		o.Namespace = namespace
	}
}
