package main

import (
	"context"
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

func setupRedisClients() (redis.Client, redis.Client){
	rdb1 := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDatabases[0],
	})

	rdb2 := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDatabases[1],
	})

	return *rdb1, *rdb2
}

// Writes data to two databases inside Redis on startup.
func setDefaultValuesOnStartup(rdb1, rdb2 redis.Client) error{
	err := rdb1.Set(ctx, "key1", "value1", 0).Err()
	if err != nil {
		return err
	}

	err = rdb2.Set(ctx, "key2", "value2", 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func main() {
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

	rdb1, rdb2 := setupRedisClients()
	if err != nil {
		zap.S().Fatal(err)
	}

	err = setDefaultValuesOnStartup(rdb1, rdb2)
	if err != nil {
		zap.S().Fatal(err)
	}

	zap.S().Info("Default values were set to both Redis databases.")

	// Create a new instance of the collector and register it with the prometheus client.
	collector := collector.NewMetricsCollector(ctx, rdb1, rdb2, cfg.RequiredMetrics)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	zap.S().Infof("Starting the server on port %s", cfg.ExporterPort)
	zap.S().Fatal(http.ListenAndServe(cfg.ExporterPort, nil))
}
