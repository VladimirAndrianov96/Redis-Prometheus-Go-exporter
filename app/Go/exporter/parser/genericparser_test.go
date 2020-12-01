package parser_test

import (
	"context"
	"exporter/exporter/client/mocks"
	"exporter/exporter/parser"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"reflect"
)

var _ = Describe("Generic INFO parser", func() {
	var (
		mockCtrl        *gomock.Controller
		ctx             context.Context
		mockClient      *mocks.MockRedisClient
		requiredMetrics []string
	)

	Describe("Requesting Redis INFO metrics", func() {
		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			ctx = context.Background()
			mockClient = mocks.NewMockRedisClient(mockCtrl)
			requiredMetrics = []string{"Clients", "Keyspace", "Memory"}
		})

		When("Metrics were fetched from Redis", func() {
			BeforeEach(func() {
				// Set up responses to be returned from mocked Redis client.
				clientsResponse := redis.NewStringResult("# Clients\nconnected_clients:3\nclient_longest_output_list:0\nclient_biggest_input_buf:0\nblocked_clients:0\n", nil)
				keyspaceResponse := redis.NewStringResult("# Keyspace\ndb1:keys=2,expires=0,avg_ttl=0\ndb2:keys=1,expires=0,avg_ttl=0\ndb3:keys=1,expires=0,avg_ttl=0\n", nil)
				memoryResponse := redis.NewStringResult("# Memory\nused_memory:862632\nused_memory_human:842.41K\nused_memory_rss:7655424\nused_memory_rss_human:7.30M\nused_memory_peak:945504\nused_memory_peak_human:923.34K\ntotal_system_memory:13347020800\ntotal_system_memory_human:12.43G\nused_memory_lua:37888\nused_memory_lua_human:37.00K\nmaxmemory:0\nmaxmemory_human:0B\nmaxmemory_policy:noeviction\nmem_fragmentation_ratio:8.87\nmem_allocator:jemalloc-4.0.3\n", nil)

				mockClient.EXPECT().Info(ctx, "Clients").Return(clientsResponse)
				mockClient.EXPECT().Info(ctx, "Keyspace").Return(keyspaceResponse)
				mockClient.EXPECT().Info(ctx, "Memory").Return(memoryResponse)
			})
			It("Returns Prometheus-formatted metrics", func() {
				res, err := parser.GetInfoMetrics(ctx, requiredMetrics, mockClient)

				Expect(err).To(BeNil())
				Expect(reflect.DeepEqual(res, getGenericExpectedData())).To(BeTrue())
			})
		})
	})
})

func getGenericExpectedData() *map[string]string {
	metrics := make(map[string]string)

	metrics["mem_fragmentation_ratio"] = "8.87"
	metrics["used_memory_rss_human"] = "7.30M"
	metrics["total_system_memory_human"] = "12.43G"
	metrics["used_memory_lua_human"] = "37.00K"
	metrics["maxmemory_human"] = "0B"
	metrics["client_longest_output_list"] = "0"
	metrics["blocked_clients"] = "0"
	metrics["used_memory_rss"] = "7655424"
	metrics["total_system_memory"] = "13347020800"
	metrics["maxmemory_policy"] = "noeviction"
	metrics["connected_clients"] = "3"
	metrics["client_biggest_input_buf"] = "0"
	metrics["used_memory_peak"] = "945504"
	metrics["used_memory_peak_human"] = "923.34K"
	metrics["mem_allocator"] = "jemalloc-4.0.3"
	metrics["used_memory"] = "862632"
	metrics["used_memory_human"] = "842.41K"
	metrics["used_memory_lua"] = "37888"
	metrics["maxmemory"] = "0"

	return &metrics
}
