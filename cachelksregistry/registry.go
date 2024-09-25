package cachelksregistry

import (
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelks"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelks/gocachelks"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelks/redislks"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-common/util/promutil"
	"github.com/rs/zerolog/log"
)

type Config struct {
	GoCache    []gocachelks.Config             `mapstructure:"go-cache,omitempty" json:"go-cache,omitempty" yaml:"go-cache,omitempty"`
	Redis      []redislks.Config               `mapstructure:"redis,omitempty" json:"redis,omitempty" yaml:"redis,omitempty"`
	MetricsCfg promutil.MetricsConfigReference `yaml:"metrics,omitempty" mapstructure:"metrics,omitempty" json:"metrics,omitempty"`
}

const DefaultCacheBrokerName = "default"

type LinkedServices []cachelks.LinkedService

var theRegistry LinkedServices

func Initialize(cfg Config) (LinkedServices, error) {

	const semLogContext = "cache-lks-registry::initialize"
	if len(theRegistry) != 0 {
		log.Warn().Msg(semLogContext + " registry already configured.. overwriting")
	}

	var r LinkedServices
	if len(cfg.Redis) > 0 {
		log.Info().Int("no-redis-linked-services", len(cfg.Redis)).Msg(semLogContext)

		for _, rcfg := range cfg.Redis {
			if rcfg.MetricsCfg.IsZero() {
				rcfg.MetricsCfg = cfg.MetricsCfg
			}
			lks, err := redislks.NewInstanceWithConfig(rcfg)
			if err != nil {
				return nil, err
			}

			r = append(r, lks)
			log.Info().Str("name", rcfg.Name).Msg(semLogContext + " redis cache instance configured")
		}
	}

	if len(cfg.GoCache) > 0 {
		log.Info().Int("no-go-cache-linked-services", len(cfg.GoCache)).Msg(semLogContext)

		for _, rcfg := range cfg.GoCache {
			if rcfg.MetricsCfg.IsZero() {
				rcfg.MetricsCfg = cfg.MetricsCfg
			}
			lks, err := gocachelks.NewInstanceWithConfig(rcfg)
			if err != nil {
				return nil, err
			}

			r = append(r, lks)
			log.Info().Str("name", rcfg.Name).Msg(semLogContext + " go-cache instance configured")
		}
	}

	if len(r) == 0 {
		log.Info().Msg(semLogContext + " no cache config provided....skipping")
	} else {
		theRegistry = r
	}

	return r, nil
}

func GetLinkedServiceOfType(typ string, name string) (cachelks.LinkedService, error) {
	for _, r := range theRegistry {
		if r.Type() == typ && r.Name() == name {
			return r, nil
		}
	}

	return nil, fmt.Errorf("cannot find cache of type %s by name [%s]", typ, name)
}
