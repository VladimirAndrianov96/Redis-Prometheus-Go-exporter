package collector

import (
	"context"
	"exporter/exporter/client"
	"exporter/exporter/parser"
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
		[]string{"database"}, nil,
	)
	expiringKeysCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "expiring_keys_count"),
		"Number of keys per Redis database.",
		[]string{"database"}, nil,
	)
	averageKeyTTLSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "average_key_ttl_seconds"),
		"Average key TTL in seconds.",
		[]string{"database"}, nil,
	)
)

type MetricsCollector struct {
	ctx                   context.Context
	clients               client.SliceOfClients
	requiredMetrics       []string
	databases             []int
	clientsConnectedTotal *prometheus.Desc
	keysPerDatabaseCount  *prometheus.Desc
	expiringKeysCount     *prometheus.Desc
	averageKeyTTLSeconds  *prometheus.Desc
}

// NewMetricsCollector allocates a new collector instance.
func NewMetricsCollector(ctx context.Context, clients client.SliceOfClients, requiredMetrics []string, databases []int) *MetricsCollector {
	return &MetricsCollector{
		ctx:                   ctx,
		clients:               clients,
		databases:             databases,
		requiredMetrics:       requiredMetrics,
		clientsConnectedTotal: clientsConnectedTotal,
		keysPerDatabaseCount:  keysPerDatabaseCount,
		expiringKeysCount:     expiringKeysCount,
		averageKeyTTLSeconds:  averageKeyTTLSeconds,
	}
}

// Describe writes all descriptors to the Prometheus desc channel.
func (collector *MetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.clientsConnectedTotal
	ch <- collector.keysPerDatabaseCount
	ch <- collector.expiringKeysCount
	ch <- collector.averageKeyTTLSeconds
}

// Collect implements required collect function for all Prometheus collectors
func (collector *MetricsCollector) Collect(ch chan<- prometheus.Metric) {
	// Any of clients from same Redis connection works well to provide collector with general and keyspace data from INFO.
	generalMetrics, err := parser.GetInfoMetrics(collector.ctx, collector.requiredMetrics, collector.clients.RedisClients[0])
	if err != nil {
		zap.S().Panic(err)
	}

	keyspaceMetrics, err := parser.GetKeyspaceMetrics(collector.ctx, collector.clients.RedisClients[0])
	if err != nil {
		zap.S().Panic(err)
	}

	// Non-numerical values cannot be set as values for Prometheus metrics.
	// Store this exceptional data and return it later as labels for metric.
	stringMetricsKeys := []string{}
	stringMetricsValues := []string{}

	// Iterate over all metrics.
	for k, v := range *generalMetrics {
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			stringMetricsKeys = append(stringMetricsKeys, k)
			stringMetricsValues = append(stringMetricsValues, v)
			continue
		}

		numericalMetric := prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "info", k),
			"Data gathered from Redis INFO.",
			nil, nil,
		)

		// Return all numerical metrics.
		ch <- prometheus.MustNewConstMetric(numericalMetric, prometheus.GaugeValue, val)
	}

	// Return all non-numeric metrics as labels.
	stringMetric := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "info", "non_numerical"),
		"Non-numerical data gathered from Redis INFO.",
		stringMetricsKeys, nil,
	)

	ch <- prometheus.MustNewConstMetric(stringMetric, prometheus.GaugeValue, 1, stringMetricsValues...)

	// Return required common custom metric.
	ch <- prometheus.MustNewConstMetric(collector.clientsConnectedTotal, prometheus.GaugeValue, getClientsConnectedTotal(*generalMetrics))

	// Return required metrics for all configured databases.
	for i, v := range *keyspaceMetrics {
		db := strconv.Itoa(collector.databases[i])
		ch <- prometheus.MustNewConstMetric(collector.keysPerDatabaseCount, prometheus.GaugeValue, getKeysPerDatabaseCount(v), db)
		ch <- prometheus.MustNewConstMetric(collector.expiringKeysCount, prometheus.GaugeValue, getExpiringKeysCount(v), db)
		ch <- prometheus.MustNewConstMetric(collector.averageKeyTTLSeconds, prometheus.GaugeValue, getAverageKeyTTLSeconds(v), db)
	}
}

func getClientsConnectedTotal(metrics map[string]string) float64 {
	metric := "connected_clients"
	val, err := strconv.ParseFloat(metrics[metric], 64)
	if err != nil {
		zap.S().Panic("Failed to read metric:", metric, err)
	}

	return val
}

func getKeysPerDatabaseCount(metrics map[string]string) float64 {
	metric := "keys"
	val, err := strconv.ParseFloat(metrics[metric], 64)
	if err != nil {
		zap.S().Panic("Failed to read metric:", metric, err)
	}

	return val
}

func getExpiringKeysCount(metrics map[string]string) float64 {
	metric := "expires"
	val, err := strconv.ParseFloat(metrics[metric], 64)
	if err != nil {
		zap.S().Panic("Failed to read metric:", metric, err)
	}

	return val
}

func getAverageKeyTTLSeconds(metrics map[string]string) float64 {
	metric := "avg_ttl"
	val, err := strconv.ParseFloat(metrics[metric], 64)
	if err != nil {
		zap.S().Panic("Failed to read metric:", metric, err)
	}

	return val
}
