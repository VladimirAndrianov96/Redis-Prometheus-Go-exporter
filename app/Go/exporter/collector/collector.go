package collector

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

const namespace = "redis"
var metrics map[string]string

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
	var metricValue float64
	if 1 == 1 {
		metricValue = 1
	}

	// Write latest value for each metric in the Prometheus metric channel.
	ch <- prometheus.MustNewConstMetric(collector.clientsConnectedTotal, prometheus.GaugeValue, metricValue)
	ch <- prometheus.MustNewConstMetric(collector.keysPerDatabaseCount, prometheus.GaugeValue, metricValue)
	ch <- prometheus.MustNewConstMetric(collector.expiringKeysCount, prometheus.GaugeValue, metricValue)
	ch <- prometheus.MustNewConstMetric(collector.averageKeyTTLSeconds, prometheus.GaugeValue, metricValue)
}

func (collector *metricsCollector) getMetrics(){
	val := collector.rdb1.Info(collector.ctx)
	fmt.Println(val)
}

func getClientsConnectedTotal(){

}

func getKeysPerDatabaseCount(){

}

func getExpiringKeysCount(){

}

func getAverageKeyTTLSecondsl(){

}