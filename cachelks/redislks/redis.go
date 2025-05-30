package redislks

import (
	"context"
	"errors"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelks"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-common/util/promutil"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/har"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type LinkedService struct {
	cfg  Config
	rdbs map[int]*redis.Client
}

func NewInstanceWithConfig(cfg Config) (*LinkedService, error) {
	lks := &LinkedService{cfg: cfg}
	return lks, nil
}

func (lks *LinkedService) Name() string {
	return lks.cfg.Name
}

func (lks *LinkedService) Type() string {
	return RedisLinkedServiceType
}

func (lks *LinkedService) Url(forPath string) string {
	ub := har.UrlBuilder{}
	//ub.WithPort(80)
	ub.WithScheme("redis")
	ub.WithHostname(lks.cfg.Addr)
	ub.WithPath(forPath)
	return ub.Url()
}

func (lks *LinkedService) getClient(aDb int) (*redis.Client, error) {

	if aDb == RedisUseLinkedServiceConfiguredIndex {
		aDb = lks.cfg.Db
	}

	if lks.rdbs == nil {
		lks.rdbs = make(map[int]*redis.Client)
	}

	rdb, ok := lks.rdbs[aDb]
	if !ok {
		rdb = redis.NewClient(&redis.Options{
			Addr:         lks.cfg.Addr,
			Password:     lks.cfg.Passwd,
			DB:           aDb,
			PoolSize:     lks.cfg.PoolSize,
			MaxRetries:   lks.cfg.MaxRetries,
			DialTimeout:  time.Duration(lks.cfg.DialTimeout) * time.Millisecond,
			ReadTimeout:  time.Duration(lks.cfg.ReadTimeout) * time.Millisecond,
			WriteTimeout: time.Duration(lks.cfg.WriteTimeout) * time.Millisecond,
			// IdleTimeout:  time.Duration(lks.cfg.IdleTimeout) * time.Millisecond,
		})
		lks.rdbs[aDb] = rdb
	}

	return rdb, nil
}

func (lks *LinkedService) Set(ctx context.Context, key string, value interface{}, opts cachelks.CacheOptions) error {
	const semLogContext = "redis-lks::set"
	beginOf := time.Now()
	lbls := lks.MetricsLabels(http.MethodPost)
	defer func(start time.Time) {
		_ = lks.setMetrics(start, lbls)
	}(beginOf)

	rdb, err := lks.getClient(RedisUseLinkedServiceConfiguredIndex)
	if err != nil {
		return err
	}

	// Check to use the specific ttl.
	dataTtl := lks.cfg.TTL
	if opts.Ttl > 0 {
		dataTtl = opts.Ttl
	}

	var sts *redis.StatusCmd
	switch tv := value.(type) {
	case []byte:
		sts = rdb.Set(ctx, key, tv, dataTtl)
	default:
		sts = rdb.Set(ctx, key, value, dataTtl)
	}

	err = sts.Err()
	if err == nil {
		lbls[MetricIdStatusCode] = "200"
	}
	return err
}

func (lks *LinkedService) Get(ctx context.Context, key string, opts cachelks.CacheOptions) (interface{}, error) {

	const semLogContext = "redis-lks::get"
	beginOf := time.Now()
	lbls := lks.MetricsLabels(http.MethodGet)
	defer func(start time.Time) {
		_ = lks.setMetrics(start, lbls)
	}(beginOf)

	rdb, err := lks.getClient(RedisUseLinkedServiceConfiguredIndex)
	if err != nil {
		return nil, err
	}

	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			log.Trace().Str("key", key).Msg(semLogContext + " cached key not found")
			lbls[MetricIdStatusCode] = fmt.Sprint(http.StatusNotFound)
			return nil, nil
		}
		return nil, err
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
		MetricIdCacheType:      RedisLinkedServiceType,
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
