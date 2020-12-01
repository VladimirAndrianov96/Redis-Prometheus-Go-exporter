package parser

import (
	"context"
	"exporter/exporter/client"
	"strings"
)

func GetKeyspaceMetrics(ctx context.Context, client client.RedisClient) (*[]map[string]string, error) {
	// Return array of maps with values per db.
	metricsForAllDB := []map[string]string{}

	// Get Redis INFO keyspace section data by querying it via client.
	data := client.Info(ctx, "Keyspace")

	// Separate plain string of values into slice of strings.
	// Fix for Windows line endings included (if ran locally in Windows).
	slicedData := strings.Split(strings.Replace(data.String(), "\r\n", "\n", -1), "\n")

	// Remove the "db1:" part from the "db1:keys:=1..." response to ease the parsing logic.
	for k, v := range slicedData {
		slicedData[k] = strings.Join(strings.Split(v, ":")[1:], ":")
	}

	// Remove "# Keyspace" info section header from output, it is always first line and
	// remove the trailing new line by dropping last element instead of iterating the whole slice.
	slicedData = slicedData[1:]
	slicedData = slicedData[:len(slicedData)-1]

	// Iterate over keyspace data for each database.
	for _, v := range slicedData {
		metrics := make(map[string]string)
		// Separate strings using comma separator.
		separatedData := strings.Split(strings.Replace(v, ",", "\n", -1), "\n")

		// Add key-value entry to the metrics map.
		for _, dataRow := range separatedData {
			// Split string by "=" delimiter to separate the string into key and value.
			parts := strings.Split(dataRow, "=")
			metrics[parts[0]] = parts[1]
		}

		// Append resulting slice with map of metrics per db.
		metricsForAllDB = append(metricsForAllDB, metrics)
	}

	return &metricsForAllDB, nil
}
