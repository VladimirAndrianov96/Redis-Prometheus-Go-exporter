package collector_test

import (
	"context"
	"exporter/exporter/client"
	"exporter/exporter/client/mocks"
	"exporter/exporter/collector"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"net/http/httptest"
)

// Test exporter with 3 databases configured.
// Exporter fetches as many databases as configured.
var _ = Describe("Redis collector Prometheus exporter", func() {
	var (
		mockCtrl         *gomock.Controller
		ctx              context.Context
		mockClients      client.SliceOfClients
		metricsCollector *collector.MetricsCollector
		mockClient1      *mocks.MockRedisClient
		mockClient2      *mocks.MockRedisClient
		mockClient3      *mocks.MockRedisClient
		handler          http.Handler
	)

	Describe("Requesting Redis metrics", func() {
		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			ctx = context.Background()

			// Set up test mocked Redis clients.
			mockClients = client.SliceOfClients{}
			mockClient1 = mocks.NewMockRedisClient(mockCtrl)
			mockClient2 = mocks.NewMockRedisClient(mockCtrl)
			mockClient3 = mocks.NewMockRedisClient(mockCtrl)
			mockClients.RedisClients = append(mockClients.RedisClients, mockClient1)
			mockClients.RedisClients = append(mockClients.RedisClients, mockClient2)
			mockClients.RedisClients = append(mockClients.RedisClients, mockClient3)

			// Set up collector to use mocked Redis clients.
			metricsCollector = collector.NewMetricsCollector(ctx, mockClients, []string{"Keyspace", "Clients", "Memory"}, []int{1, 2, 3})

			// Get rid of any additional metrics, it should expose only required metrics with a custom registry
			r := prometheus.NewRegistry()
			r.MustRegister(metricsCollector)
			handler = promhttp.HandlerFor(r, promhttp.HandlerOpts{})
		})

		When("Metrics were fetched from Redis", func() {
			BeforeEach(func() {
				// Set up responses to be returned from mocked Redis client.
				clientsResponse := redis.NewStringResult("# Clients\nconnected_clients:3\nclient_longest_output_list:0\nclient_biggest_input_buf:0\nblocked_clients:0\n", nil)
				keyspaceResponse := redis.NewStringResult("# Keyspace\ndb1:keys=2,expires=0,avg_ttl=0\ndb2:keys=1,expires=0,avg_ttl=0\ndb3:keys=1,expires=0,avg_ttl=0\n", nil)
				memoryResponse := redis.NewStringResult("# Memory\nused_memory:862632\nused_memory_human:842.41K\nused_memory_rss:7655424\nused_memory_rss_human:7.30M\nused_memory_peak:945504\nused_memory_peak_human:923.34K\ntotal_system_memory:13347020800\ntotal_system_memory_human:12.43G\nused_memory_lua:37888\nused_memory_lua_human:37.00K\nmaxmemory:0\nmaxmemory_human:0B\nmaxmemory_policy:noeviction\nmem_fragmentation_ratio:8.87\nmem_allocator:jemalloc-4.0.3\n", nil)

				mockClient1.EXPECT().Info(ctx, "Clients").Return(clientsResponse)
				mockClient1.EXPECT().Info(ctx, "Keyspace").Return(keyspaceResponse)
				mockClient1.EXPECT().Info(ctx, "Memory").Return(memoryResponse)

			})
			It("Returns Prometheus-formatted metrics", func() {
				req, err := http.NewRequest("GET", "/metrics", nil)
				Expect(err).To(BeNil())

				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, req)

				Expect(rr.Body.String()).To(Equal(getExpectedData()))
				Expect(rr.Code).To(Equal(http.StatusOK))
			})
		})
	})
})

func getExpectedData() string {
	return `# HELP redis_average_key_ttl_seconds Average key TTL in seconds.
# TYPE redis_average_key_ttl_seconds gauge
redis_average_key_ttl_seconds{database="1"} 0
redis_average_key_ttl_seconds{database="2"} 0
redis_average_key_ttl_seconds{database="3"} 0
# HELP redis_clients_connected_total Total number of clients connected to Redis.
# TYPE redis_clients_connected_total gauge
redis_clients_connected_total 3
# HELP redis_expiring_keys_count Number of keys per Redis database.
# TYPE redis_expiring_keys_count gauge
redis_expiring_keys_count{database="1"} 0
redis_expiring_keys_count{database="2"} 0
redis_expiring_keys_count{database="3"} 0
# HELP redis_info_blocked_clients Data gathered from Redis INFO.
# TYPE redis_info_blocked_clients gauge
redis_info_blocked_clients 0
# HELP redis_info_client_biggest_input_buf Data gathered from Redis INFO.
# TYPE redis_info_client_biggest_input_buf gauge
redis_info_client_biggest_input_buf 0
# HELP redis_info_client_longest_output_list Data gathered from Redis INFO.
# TYPE redis_info_client_longest_output_list gauge
redis_info_client_longest_output_list 0
# HELP redis_info_connected_clients Data gathered from Redis INFO.
# TYPE redis_info_connected_clients gauge
redis_info_connected_clients 3
# HELP redis_info_maxmemory Data gathered from Redis INFO.
# TYPE redis_info_maxmemory gauge
redis_info_maxmemory 0
# HELP redis_info_mem_fragmentation_ratio Data gathered from Redis INFO.
# TYPE redis_info_mem_fragmentation_ratio gauge
redis_info_mem_fragmentation_ratio 8.87
# HELP redis_info_non_numerical Non-numerical data gathered from Redis INFO.
# TYPE redis_info_non_numerical gauge
redis_info_non_numerical{maxmemory_human="0B",maxmemory_policy="noeviction",mem_allocator="jemalloc-4.0.3",total_system_memory_human="12.43G",used_memory_human="842.41K",used_memory_lua_human="37.00K",used_memory_peak_human="923.34K",used_memory_rss_human="7.30M"} 1
# HELP redis_info_total_system_memory Data gathered from Redis INFO.
# TYPE redis_info_total_system_memory gauge
redis_info_total_system_memory 1.33470208e+10
# HELP redis_info_used_memory Data gathered from Redis INFO.
# TYPE redis_info_used_memory gauge
redis_info_used_memory 862632
# HELP redis_info_used_memory_lua Data gathered from Redis INFO.
# TYPE redis_info_used_memory_lua gauge
redis_info_used_memory_lua 37888
# HELP redis_info_used_memory_peak Data gathered from Redis INFO.
# TYPE redis_info_used_memory_peak gauge
redis_info_used_memory_peak 945504
# HELP redis_info_used_memory_rss Data gathered from Redis INFO.
# TYPE redis_info_used_memory_rss gauge
redis_info_used_memory_rss 7.655424e+06
# HELP redis_keys_per_database_count Number of keys per Redis database.
# TYPE redis_keys_per_database_count gauge
redis_keys_per_database_count{database="1"} 2
redis_keys_per_database_count{database="2"} 1
redis_keys_per_database_count{database="3"} 1
`
}
