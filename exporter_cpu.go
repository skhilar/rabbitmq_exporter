package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

var (
	utilizationDesc = prometheus.NewDesc(
		"rabbitmq:cpu_utilization:rate5m",
		"rabbitmq_exporter: CPU utilization",
		[]string{"hsdp_instance_guid", "hsdp_instance_node_name"},
		nil,
	)
)

func init() {
	fmt.Println("registering cpu exporter")
	RegisterExporter("cpu", newExporterCPU)
}

type exporterCPU struct {
	client       *Client
	namespace    string
	resourceId   string
	instanceGuid string
}

func newExporterCPU() Exporter {
	return exporterCPU{client: NewClient(config.PrometheusHost, config.PrometheusPort),
		namespace: config.ServiceNameSpace, resourceId: config.ResourceID, instanceGuid: config.ServiceInstanceGUID}
}

func (e exporterCPU) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		query := fmt.Sprintf("query=sum(rate(container_cpu_usage_seconds_total{namespace=\"%s\", container=\"%s\", pod=~\"rmq-%s.*\"}[5m])) by (node, container)", e.namespace, "rabbitmq-k8s", e.resourceId)
		resp, err := e.client.execute("GET", "/api/v1/query", query)
		if err != nil {
			fmt.Println(err)
			return
		}
		if len(resp.Data.Result) == 0 {
			return
		}
		for _, r := range resp.Data.Result {
			val, err := getFloat(r.Value[1])
			if err != nil {
				return
			}
			ch <- prometheus.MustNewConstMetric(utilizationDesc,
				prometheus.GaugeValue,
				val,
				e.instanceGuid,
				r.Metric.Node,
			)

		}
	}()
	wg.Wait()
	return nil
}

func (e exporterCPU) Describe(ch chan<- *prometheus.Desc) {
	ch <- utilizationDesc
}
