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

func setupRedisClients() ([]*redis.Client){
	clients := []*redis.Client{}

	for i, _ := range cfg.RedisDatabases{
		client := redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddress,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDatabases[i],
		})

		clients = append(clients, client)
	}

	return clients
}

// ***COMMENTED OUT FOR DEMO PURPOSES***
// Writes data to two databases inside Redis on startup.
//func setDefaultValuesOnStartup(clients []*redis.Client) error{
//	err := clients[0].Set(ctx, "key1", "value1", 0).Err()
//	if err != nil {
//		return err
//	}
//
//	err = clients[1].Set(ctx, "key2", "value2", 0).Err()
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

// Writes data to all configured Redis databases on startup to make Redis create them.
// If there are 5 databases configured, then create 5 and get all metrics.
func setDefaultValuesOnStartup(clients []*redis.Client) error{
	// Add more data to two first database to make difference in metrics noticeable.
	err := clients[0].Set(ctx, "test", "test", 0).Err()
	if err != nil {
		return err
	}

	for i, v := range clients{
		index := string(i)
		err := v.Set(ctx, "key"+index, "value"+index, 0).Err()
		if err != nil {
			return err
		}
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

	clients := setupRedisClients()

	err = setDefaultValuesOnStartup(clients)
	if err != nil {
		zap.S().Fatal(err)
	}

	zap.S().Info("Default values were set to both Redis databases.")
	// Create a new instance of the collector and register it with the prometheus client.
	collector := collector.NewMetricsCollector(ctx, clients, cfg.RequiredMetrics, cfg.RedisDatabases)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	zap.S().Infof("Starting the server on port %s", cfg.ExporterPort)
	zap.S().Fatal(http.ListenAndServe(cfg.ExporterPort, nil))
}
