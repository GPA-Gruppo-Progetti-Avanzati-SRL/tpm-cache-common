package gocachelks

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-common/util/promutil"
	"time"
)

const (
	GoCacheLinkedServiceType = "go-cache"
)

type Config struct {
	Name              string                          `mapstructure:"name,omitempty" json:"name,omitempty" yaml:"name,omitempty"`
	DefaultExpiration time.Duration                   `yaml:"default-expiration,omitempty" mapstructure:"default-expiration,omitempty" json:"default-expiration,omitempty"`
	CleanupInterval   time.Duration                   `yaml:"cleanup-interval,omitempty" mapstructure:"cleanup-interval,omitempty" json:"cleanup-interval,omitempty"`
	MetricsCfg        promutil.MetricsConfigReference `yaml:"metrics,omitempty" mapstructure:"metrics,omitempty" json:"metrics,omitempty"`
}

func (c *Config) PostProcess() error {
	return nil
}
