package collector

import (
	"go.uber.org/zap"
	"strings"
)

var metrics map[string]string

func (collector *metricsCollector) getInfoMetrics(){
	metrics = make(map[string]string)

	// Iterate over passed sections.
	for _, section := range collector.requiredMetrics{
		data, err := collector.rdb1.Info(collector.ctx, section).Result()
		if err != nil {
			zap.S().Panic(err)
		}

		// Separate plain string of values into slice of strings.
		// Fix for Windows line endings included (if ran locally in Windows).
		slicedData := strings.Split(strings.Replace(data, "\r\n", "\n", -1), "\n")

		// Remove "# Clients" info section header from output, it is always first line and
		// remove the trailing new line by dropping last element instead of iterating the whole slice.
		slicedData = slicedData[1:]
		slicedData = slicedData[:len(slicedData)-1]

		// Add key-value entry to the metrics map.
		for _, dataRow := range slicedData{
			parts := strings.Split(dataRow, ":")
			metrics[parts[0]]=parts[1]
		}
	}
}