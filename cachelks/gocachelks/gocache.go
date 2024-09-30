package gocachelks

import (
	"context"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelks"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelks/gocachelks/gocache"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-common/util/promutil"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/har"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type LinkedService struct {
	cfg   Config
	cache *gocache.Cache
}

func NewInstanceWithConfig(cfg Config) (*LinkedService, error) {
	lks := &LinkedService{cfg: cfg, cache: gocache.New(cfg.DefaultExpiration, cfg.CleanupInterval)}
	return lks, nil
}

func (lks *LinkedService) Name() string {
	return lks.cfg.Name
}

func (lks *LinkedService) Type() string {
	return GoCacheLinkedServiceType
}

func (lks *LinkedService) Size() int {
	return lks.cache.ItemCount()
}

func (lks *LinkedService) Items() map[string]interface{} {
	if lks.Size() == 0 {
		return nil
	}

	items := make(map[string]interface{}, lks.cache.ItemCount())
	for n, v := range lks.cache.Items() {
		items[n] = v.Object
	}

	return items
}

func (lks *LinkedService) Url(forPath string) string {
	ub := har.UrlBuilder{}
	ub.WithPort(1111)
	ub.WithScheme("go-cache")
	ub.WithHostname("localhost")
	ub.WithPath(forPath)
	return ub.Url()
}

func (lks *LinkedService) Set(ctx context.Context, key string, value interface{}, opts cachelks.CacheOptions) error {
	const semLogContext = "go-cache-lks::set"
	beginOf := time.Now()
	lbls := lks.MetricsLabels(http.MethodPost)
	defer func(start time.Time) {
		_ = lks.setMetrics(start, lbls)
	}(beginOf)

	// Check to use the specific ttl.
	dataTtl := lks.cfg.DefaultExpiration
	if opts.Ttl > 0 {
		dataTtl = opts.Ttl
	}

	switch tv := value.(type) {
	case []byte:
		lks.cache.Set(key, tv, dataTtl)
	default:
		lks.cache.Set(key, value, dataTtl)
	}

	lbls[MetricIdStatusCode] = "200"
	return nil
}

func (lks *LinkedService) Get(ctx context.Context, key string, opts cachelks.CacheOptions) (interface{}, error) {

	const semLogContext = "redis-lks::get"
	beginOf := time.Now()
	lbls := lks.MetricsLabels(http.MethodGet)
	defer func(start time.Time) {
		_ = lks.setMetrics(start, lbls)
	}(beginOf)

	val, found := lks.cache.Get(key)
	if !found {
		log.Warn().Str("key", key).Msg(semLogContext + " cached key not found")
		lbls[MetricIdStatusCode] = fmt.Sprint(http.StatusNotFound)
		return nil, nil
	}

	lbls[MetricIdStatusCode] = fmt.Sprint(http.StatusOK)
	return val, nil
}

const (
	MetricIdStatusCode     = "status-code"
	MetricIdCacheOperation = "operation"
	MetricIdCacheType      = "cache-type"
)

func (lks *LinkedService) MetricsLabels(m string) prometheus.Labels {

	metricsLabels := prometheus.Labels{
		MetricIdStatusCode:     fmt.Sprint(http.StatusInternalServerError),
		MetricIdCacheOperation: m,
		MetricIdCacheType:      GoCacheLinkedServiceType,
	}

	return metricsLabels
}

func (lks *LinkedService) setMetrics(begin time.Time, lbls prometheus.Labels) error {
	const semLogContext = "redis-lks::set-metrics"

	var err error
	var g promutil.Group

	cfg := lks.cfg.MetricsCfg
	if cfg.GId != "" && cfg.IsEnabled() {
		g, _, err = cfg.ResolveGroup(nil)
		if err != nil {
			log.Error().Err(err).Msg(semLogContext)
			return err
		}

		if cfg.IsCounterEnabled() {
			err = g.SetMetricValueById(cfg.CounterId, 1, lbls)
		}

		if cfg.IsHistogramEnabled() {
			err = g.SetMetricValueById(cfg.HistogramId, time.Since(begin).Seconds(), lbls)
		}
	}

	return err
}
