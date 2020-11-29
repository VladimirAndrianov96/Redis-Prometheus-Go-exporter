package collector

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"strconv"
)

const namespace = "redis"

var (
	// Metrics
	clientsConnectedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "clients_connected_total"),
		"Total number of clients connected to Redis.",
		nil, nil,
	)
	keysPerDatabaseCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "keys_per_database_count"),
		"Number of keys per Redis database.",
		nil, nil,
	)
	expiringKeysCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "expiring_keys_count"),
		"Number of keys per Redis database.",
		nil, nil,
	)
	averageKeyTTLSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "average_key_ttl_seconds"),
		"Average key TTL in seconds.",
		nil, nil,
	)
)

type metricsCollector struct {
	ctx context.Context
	rdb1 redis.Client
	rdb2 redis.Client
	requiredMetrics []string
	clientsConnectedTotal *prometheus.Desc
	keysPerDatabaseCount *prometheus.Desc
	expiringKeysCount *prometheus.Desc
	averageKeyTTLSeconds *prometheus.Desc
}

// NewMetricsCollector allocates a new collector instance.
func NewMetricsCollector(ctx context.Context, rdb1, rdb2 redis.Client, requiredMetrics []string) *metricsCollector{
	return &metricsCollector{
		ctx: ctx,
		rdb1: rdb1,
		rdb2: rdb2,
		requiredMetrics: requiredMetrics,
		clientsConnectedTotal: clientsConnectedTotal,
		keysPerDatabaseCount: keysPerDatabaseCount,
		expiringKeysCount: expiringKeysCount,
		averageKeyTTLSeconds: averageKeyTTLSeconds,
	}
}

// Describe writes all descriptors to the Prometheus desc channel.
func (collector *metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.clientsConnectedTotal
	ch <- collector.keysPerDatabaseCount
	ch <- collector.expiringKeysCount
	ch <- collector.averageKeyTTLSeconds
}

// Collect implements required collect function for all Prometheus collectors
func (collector *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	collector.getInfoMetrics()

	stringMetricsKeys := []string{}
	stringMetricsValues := []string{}

	for k, v := range metrics{
		val, err := strconv.ParseFloat(v, 64)
		if err != nil{
			stringMetricsKeys = append(stringMetricsKeys, k)
			stringMetricsValues = append(stringMetricsValues, v)
			continue
		}

		metric := prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "info", k),
			"Data gathered from Redis INFO.",
			nil, nil,
		)

		ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, val)
	}
		metric := prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "info", "non_numerical"),
			"Non-numerical data gathered from Redis INFO.",
			stringMetricsKeys, nil,
		)

		ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, 1, stringMetricsValues...)

	// Write latest value for each metric in the Prometheus metric channel.
	ch <- prometheus.MustNewConstMetric(collector.clientsConnectedTotal, prometheus.GaugeValue, getClientsConnectedTotal())
	//ch <- prometheus.MustNewConstMetric(collector.keysPerDatabaseCount, prometheus.GaugeValue, 1)
	//ch <- prometheus.MustNewConstMetric(collector.expiringKeysCount, prometheus.GaugeValue, 1)
	//ch <- prometheus.MustNewConstMetric(collector.averageKeyTTLSeconds, prometheus.GaugeValue, 1)
}

func getClientsConnectedTotal() float64{
	metric := "connected_clients"
	val, err := strconv.ParseFloat(metrics[metric], 64)
	if err != nil{
		zap.S().Panic("Failed to read %s metric", metric)
	}

	return val
}

func getKeysPerDatabaseCount(){
}

func getExpiringKeysCount(){
}

func getAverageKeyTTLSecondsl(){
}