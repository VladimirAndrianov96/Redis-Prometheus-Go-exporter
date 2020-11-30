package main

import (
	"context"
	"crypto/tls"
	"exporter/exporter/client"
	"exporter/exporter/collector"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
)

// config declares connection and parser details.
type config struct {
	ExporterPort  string `mapstructure:"exporter_port"`

	RedisAddress     string `mapstructure:"redis_address"`
	RedisPassword 	 string `mapstructure:"redis_password"`

	RedisDatabases []int `mapstructure:"redis_databases"`

	RequiredMetrics []string `mapstructure:"required_metrics"`
}

var cfg config
var ctx = context.Background()

// Initialize logger to replace the default one.
func initLogger() error {
	// Initialize the logs encoder.
	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder.EncodeDuration = zapcore.StringDurationEncoder

	// Initialize the logger.
	logger, err := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Encoding:         "console",
		EncoderConfig:    encoder,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()
	if err != nil {
		return err
	}

	// Replace the default logger with zap logger.
	zap.ReplaceGlobals(logger)

	return nil
}

// Load all configuration details from local .yaml file.
func loadConfiguration() error {
	viper.AddConfigPath("./exporter/cmd/config")
	viper.SetConfigName("configuration")

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return err
	}

	return nil
}

func setupRedisClients() (client.SliceOfClients){
	clients := client.SliceOfClients{}

	for i, _ := range cfg.RedisDatabases{
		client := redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddress,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDatabases[i],
		})

		clients.RedisClients = append(clients.RedisClients, *client)
	}

	return clients
}

// Writes data to all configured Redis databases on startup to make Redis create them.
// If there are 5 databases configured, then create 5 and get all metrics.
func setDefaultValuesOnStartup(clients client.SliceOfClients) error{
	// Add more data to two first database to make difference in metrics noticeable.
	err := clients.RedisClients[0].Set(ctx, "test", "test", 0).Err()
	if err != nil {
		return err
	}

	for i, v := range clients.RedisClients{
		index := string(i)
		err := v.Set(ctx, "key"+index, "value"+index, 0).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// Disable cert verification to use self-signed certificates for internal service needs.
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Global logging synchronizer.
	// This ensures the logged data is flushed out of the buffer before program exits.
	defer zap.S().Sync()

	err := initLogger()
	if err != nil {
		zap.S().Fatal(err)
	}

	err = loadConfiguration()
	if err != nil {
		zap.S().Fatal(err)
	}

	clients := setupRedisClients()

	err = setDefaultValuesOnStartup(clients)
	if err != nil {
		zap.S().Fatal(err)
	}

	zap.S().Info("Default values were set to both Redis databases.")

	// Create a new instance of the collector and register it with the prometheus client.
	collector := collector.NewMetricsCollector(ctx, clients, cfg.RequiredMetrics, cfg.RedisDatabases)

	// Get rid of any additional metrics, it should expose only required metrics with a custom registry
	r := prometheus.NewRegistry()
	r.MustRegister(collector)
	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)

	zap.S().Infof("Starting the server on port %s", cfg.ExporterPort)
	zap.S().Fatal(http.ListenAndServeTLS(
		cfg.ExporterPort,
		"./exporter/crt.crt",
		"./exporter/key.key",
		nil))
}
