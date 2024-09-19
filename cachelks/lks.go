package cachelks

import (
	"context"
)

type CacheLinkedServiceRef struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty"`
	Typ  string `json:"type,omitempty" yaml:"type,omitempty" mapstructure:"type,omitempty"`
}

func (edcc *CacheLinkedServiceRef) IsZero() bool {
	return edcc.Typ == "" && edcc.Name == ""
}

type LinkedService interface {
	Name() string
	Type() string
	Set(ctx context.Context, key string, value interface{}, opts ...CacheOption) error
	Get(ctx context.Context, key string, opts ...CacheOption) (interface{}, error)
	Url(forPath string) string
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
