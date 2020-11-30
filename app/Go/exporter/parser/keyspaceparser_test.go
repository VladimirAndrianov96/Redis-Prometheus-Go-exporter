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

var _ = Describe("Keyspace INFO parser", func() {
	var(
		mockCtrl *gomock.Controller
		ctx context.Context
		mockClient *mocks.MockRedisClient
	)

	Describe("Requesting Redis INFO metrics", func() {
		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			ctx = context.Background()
			mockClient = mocks.NewMockRedisClient(mockCtrl)
		})

		When("Metrics were fetched from Redis", func() {
			BeforeEach(func(){
				// Set up responses to be returned from mocked Redis client.
				keyspaceResponse := redis.NewStringResult("# Keyspace\ndb1:keys=2,expires=0,avg_ttl=0\ndb2:keys=1,expires=0,avg_ttl=0\ndb3:keys=1,expires=0,avg_ttl=0\n", nil)

				mockClient.EXPECT().Info(ctx, "Keyspace").Return(keyspaceResponse)
			})
			It("Returns Prometheus-formatted metrics", func() {
				res, err := parser.GetKeyspaceMetrics(ctx, mockClient)

				Expect(err).To(BeNil())
				Expect(reflect.DeepEqual(res, getKeyspaceExpectedData())).To(BeTrue())
			})
		})
	})
})

func getKeyspaceExpectedData() *[]map[string]string{
	metricsForAllDB := []map[string]string{}

	metrics := make(map[string]string)
	metrics["avg_ttl"]="0"
	metrics["expires"]="0"
	metrics["keys"]="2"
	metricsForAllDB = append(metricsForAllDB, metrics)

	metrics = make(map[string]string)
	metrics["avg_ttl"]="0"
	metrics["expires"]="0"
	metrics["keys"]="1"
	metricsForAllDB = append(metricsForAllDB, metrics)

	metrics = make(map[string]string)
	metrics["avg_ttl"]="0"
	metrics["expires"]="0"
	metrics["keys"]="1"
	metricsForAllDB = append(metricsForAllDB, metrics)

	return &metricsForAllDB
}