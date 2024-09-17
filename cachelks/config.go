package cachelks

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/redislks"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-common/util/promutil"
)

type Config struct {
	Redis      []redislks.Config               `mapstructure:"redis,omitempty" json:"redis,omitempty" yaml:"redis,omitempty"`
	MetricsCfg promutil.MetricsConfigReference `yaml:"metrics,omitempty" mapstructure:"metrics,omitempty" json:"metrics,omitempty"`
}
