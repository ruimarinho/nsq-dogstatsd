package collector

import (
	"fmt"
	"reflect"

	log "github.com/Sirupsen/logrus"
	"github.com/ruimarinho/nsq-dogstatsd/producer"
)

// Metric holds a statistical metric from nsqd.
type Metric struct {
	Name  string
	Rate  float64
	Tags  []string
	Type  string
	Value float64
}

func push(m Metric, metrics *[]Metric) {
	*metrics = append(*metrics, m)
}

func gauge(name string, value interface{}, tags []string) Metric {
	var metric Metric

	switch value.(type) {
	case bool:
		if value.(bool) {
			metric = NewGauge(name, 1, tags)
		} else {
			metric = NewGauge(name, 0, tags)
		}
	case int:
		metric = NewGauge(name, float64(value.(int)), tags)
	case int32:
		metric = NewGauge(name, float64(value.(int32)), tags)
	case int64:
		metric = NewGauge(name, float64(value.(int64)), tags)
	case uint64:
		metric = NewGauge(name, float64(value.(uint64)), tags)
	default:
		log.WithFields(log.Fields{"type": fmt.Sprintf("%s", reflect.TypeOf(value))}).Panic("unknown metric type")
	}

	log.WithFields(log.Fields{
		"name":  metric.Name,
		"value": metric.Value,
		"type":  metric.Type,
		"tags":  metric.Tags,
		"rate":  metric.Rate,
	}).Info("collected metric")

	return metric
}

func NewGauge(metric string, value float64, tags []string) Metric {
	return Metric{
		Name:  metric,
		Value: value,
		Type:  "gauge",
		Tags:  tags,
		Rate:  1,
	}
}

func CollectMetrics(producer producer.Producer) ([]Metric, error) {
	log.WithField("node", producer.Hostname).Infof(`collecting metrics for node %s`, producer.Hostname)

	stats, err := producer.GetStats()
	if err != nil {
		return nil, err
	}

	var metrics []Metric
	push(gauge("topic.count", len(stats.Data.Topics), producer.GetTags()), &metrics)
	push(gauge("memory.heap_objects", int64(stats.Data.Memory.HeapObjects), producer.GetTags()), &metrics)
	push(gauge("memory.heap_idle_bytes", int64(stats.Data.Memory.HeapIdleBytes), producer.GetTags()), &metrics)
	push(gauge("memory.heap_in_use_bytes", int64(stats.Data.Memory.HeapInUseBytes), producer.GetTags()), &metrics)
	push(gauge("memory.heap_released_bytes", int64(stats.Data.Memory.HeapReleasedBytes), producer.GetTags()), &metrics)
	push(gauge("memory.gc_pause_usec_100", int64(stats.Data.Memory.GCPauseUsec100), producer.GetTags()), &metrics)
	push(gauge("memory.gc_pause_usec_99", int64(stats.Data.Memory.GCPauseUsec99), producer.GetTags()), &metrics)
	push(gauge("memory.gc_pause_usec_95", int64(stats.Data.Memory.GCPauseUsec95), producer.GetTags()), &metrics)
	push(gauge("memory.next_gc_bytes", int64(stats.Data.Memory.NextGCBytes), producer.GetTags()), &metrics)
	push(gauge("memory.gc_runs", int64(stats.Data.Memory.GCTotalRuns), producer.GetTags()), &metrics)

	for _, topic := range stats.Data.Topics {
		topicTags := append([]string{}, producer.GetTags()...)
		topicTags = append(topicTags, []string{fmt.Sprintf("topic:%s", topic.TopicName)}...)

		push(gauge("topic.channels", len(topic.Channels), topicTags), &metrics)
		push(gauge("topic.depth", topic.Depth, topicTags), &metrics)
		push(gauge("topic.backend_depth", topic.BackendDepth, topicTags), &metrics)
		push(gauge("topic.messages", topic.MessageCount, topicTags), &metrics)
		push(gauge("topic.paused", topic.Paused, topicTags), &metrics)

		if topic.E2eProcessingLatency != nil {
			for _, percentile := range topic.E2eProcessingLatency.Percentiles {
				push(gauge(fmt.Sprintf("topic.e2e_processing_latency_%f", percentile["quantile"]), percentile["value"], topicTags), &metrics)
			}
		}

		for _, channel := range topic.Channels {
			channelTags := append([]string{}, topicTags...)
			channelTags = append(channelTags, []string{fmt.Sprintf("channel:%s", channel.ChannelName)}...)

			push(gauge("channel.depth", channel.Depth, channelTags), &metrics)
			push(gauge("channel.backend_depth", channel.BackendDepth, channelTags), &metrics)
			push(gauge("channel.in_flight", channel.InFlightCount, channelTags), &metrics)
			push(gauge("channel.deferred", channel.DeferredCount, channelTags), &metrics)
			push(gauge("channel.messages", channel.MessageCount, channelTags), &metrics)
			push(gauge("channel.requeued", channel.RequeueCount, channelTags), &metrics)
			push(gauge("channel.timed_out", channel.TimeoutCount, channelTags), &metrics)
			push(gauge("channel.clients", len(channel.Clients), channelTags), &metrics)
			push(gauge("channel.paused", channel.Paused, channelTags), &metrics)

			if channel.E2eProcessingLatency != nil {
				for _, percentile := range channel.E2eProcessingLatency.Percentiles {
					push(gauge(fmt.Sprintf("channel.e2e_processing_latency_%f", percentile["quantile"]), percentile["value"], channelTags), &metrics)
				}
			}

			for _, client := range channel.Clients {
				clientTags := append([]string{}, channelTags...)
				clientTags = append(clientTags, []string{
					fmt.Sprintf("client_id:%s", client.ClientID),
					fmt.Sprintf("client_agent:%s", client.UserAgent),
					fmt.Sprintf("client_hostname:%s", client.Hostname),
					fmt.Sprintf("client_address:%s", client.RemoteAddress)}...,
				)

				push(gauge("client.state", client.State, clientTags), &metrics)
				push(gauge("client.ready_count", client.ReadyCount, clientTags), &metrics)
				push(gauge("client.in_flight", client.InFlightCount, clientTags), &metrics)
				push(gauge("client.messages", client.MessageCount, clientTags), &metrics)
				push(gauge("client.finished", client.FinishCount, clientTags), &metrics)
				push(gauge("client.requeued", client.RequeueCount, clientTags), &metrics)
			}
		}
	}

	log.WithFields(log.Fields{"node": producer.Hostname}).Infof(`collected metrics for node %s`, producer.Hostname)

	return metrics, nil
}
