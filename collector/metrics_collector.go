package collector

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/ruimarinho/nsq-dogstatsd/producer"
	log "github.com/sirupsen/logrus"
)

// Metric holds a statistical metric from nsqd.
type Metric struct {
	Name  string
	Rate  float64
	Tags  []string
	Type  string
	Value float64
}

type Collector struct {
	Producer        producer.Producer
	ExcludedMetrics []*regexp.Regexp
}

func NewMetric(metric string, value float64, tags []string) Metric {
	return Metric{
		Name:  metric,
		Value: value,
		Type:  "gauge",
		Tags:  tags,
		Rate:  1,
	}
}

func NewCollector(producer producer.Producer, excludedMetrics []*regexp.Regexp) *Collector {
	return &Collector{Producer: producer, ExcludedMetrics: excludedMetrics}
}

func (c *Collector) NewGauge(name string, value interface{}, extraTags []string) Metric {
	tags := append(c.Producer.GetTags(), extraTags...)

	for _, filter := range c.ExcludedMetrics {
		if filter.MatchString(name) {
			log.Debugf("skipping metric %s", name)
			return Metric{}
		}
	}

	var metric Metric

	switch value.(type) {
	case bool:
		if value.(bool) {
			metric = NewMetric(name, 1, tags)
		} else {
			metric = NewMetric(name, 0, tags)
		}
	case int:
		metric = NewMetric(name, float64(value.(int)), tags)
	case int32:
		metric = NewMetric(name, float64(value.(int32)), tags)
	case int64:
		metric = NewMetric(name, float64(value.(int64)), tags)
	case uint64:
		metric = NewMetric(name, float64(value.(uint64)), tags)
	case float32:
		metric = NewMetric(name, float64(value.(float32)), tags)
	case float64:
		metric = NewMetric(name, value.(float64), tags)
	default:
		log.WithFields(log.Fields{"type": fmt.Sprintf("%s", reflect.TypeOf(value))}).Error("unknown metric type")
		return metric
	}

	log.WithFields(log.Fields{
		"name":  metric.Name,
		"value": metric.Value,
		"type":  metric.Type,
		"tags":  metric.Tags,
		"rate":  metric.Rate,
	}).Debugf("collecting metric %s", metric.Name)

	return metric
}

func (c *Collector) CollectMetrics() ([]Metric, error) {
	log.WithField("node", c.Producer.Hostname).Debugf(`collecting metrics for node %s`, c.Producer.Hostname)

	stats, err := c.Producer.GetStats()
	if err != nil {
		return nil, err
	}

	var metrics []Metric
	metrics = append(metrics, c.NewGauge("topic.count", len(stats.Data.Topics), []string{}))
	metrics = append(metrics, c.NewGauge("memory.heap_objects", int64(stats.Data.Memory.HeapObjects), []string{}))
	metrics = append(metrics, c.NewGauge("memory.heap_idle_bytes", int64(stats.Data.Memory.HeapIdleBytes), []string{}))
	metrics = append(metrics, c.NewGauge("memory.heap_in_use_bytes", int64(stats.Data.Memory.HeapInUseBytes), []string{}))
	metrics = append(metrics, c.NewGauge("memory.heap_released_bytes", int64(stats.Data.Memory.HeapReleasedBytes), []string{}))
	metrics = append(metrics, c.NewGauge("memory.gc_pause_usec_100", int64(stats.Data.Memory.GCPauseUsec100), []string{}))
	metrics = append(metrics, c.NewGauge("memory.gc_pause_usec_99", int64(stats.Data.Memory.GCPauseUsec99), []string{}))
	metrics = append(metrics, c.NewGauge("memory.gc_pause_usec_95", int64(stats.Data.Memory.GCPauseUsec95), []string{}))
	metrics = append(metrics, c.NewGauge("memory.next_gc_bytes", int64(stats.Data.Memory.NextGCBytes), []string{}))
	metrics = append(metrics, c.NewGauge("memory.gc_runs", int64(stats.Data.Memory.GCTotalRuns), []string{}))

	for _, topic := range stats.Data.Topics {
		topicTags := []string{fmt.Sprintf("topic:%s", topic.TopicName)}

		metrics = append(metrics, c.NewGauge("topic.channels", len(topic.Channels), topicTags))
		metrics = append(metrics, c.NewGauge("topic.depth", topic.Depth, topicTags))
		metrics = append(metrics, c.NewGauge("topic.backend_depth", topic.BackendDepth, topicTags))
		metrics = append(metrics, c.NewGauge("topic.messages", topic.MessageCount, topicTags))
		metrics = append(metrics, c.NewGauge("topic.paused", topic.Paused, topicTags))

		if topic.E2eProcessingLatency != nil {
			for _, percentile := range topic.E2eProcessingLatency.Percentiles {
				metrics = append(metrics, c.NewGauge(fmt.Sprintf("topic.e2e_processing_latency_%f", percentile["quantile"]), percentile["value"], topicTags))
			}
		}

		for _, channel := range topic.Channels {
			channelTags := append([]string{}, topicTags...)
			channelTags = append(channelTags, []string{fmt.Sprintf("channel:%s", channel.ChannelName)}...)

			metrics = append(metrics, c.NewGauge("channel.depth", channel.Depth, channelTags))
			metrics = append(metrics, c.NewGauge("channel.backend_depth", channel.BackendDepth, channelTags))
			metrics = append(metrics, c.NewGauge("channel.in_flight", channel.InFlightCount, channelTags))
			metrics = append(metrics, c.NewGauge("channel.deferred", channel.DeferredCount, channelTags))
			metrics = append(metrics, c.NewGauge("channel.messages", channel.MessageCount, channelTags))
			metrics = append(metrics, c.NewGauge("channel.requeued", channel.RequeueCount, channelTags))
			metrics = append(metrics, c.NewGauge("channel.timed_out", channel.TimeoutCount, channelTags))
			metrics = append(metrics, c.NewGauge("channel.clients", len(channel.Clients), channelTags))
			metrics = append(metrics, c.NewGauge("channel.paused", channel.Paused, channelTags))

			if channel.E2eProcessingLatency != nil {
				for _, percentile := range channel.E2eProcessingLatency.Percentiles {
					metrics = append(metrics, c.NewGauge(fmt.Sprintf("channel.e2e_processing_latency_%f", percentile["quantile"]), percentile["value"], channelTags))
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

				metrics = append(metrics, c.NewGauge("client.state", client.State, clientTags))
				metrics = append(metrics, c.NewGauge("client.ready_count", client.ReadyCount, clientTags))
				metrics = append(metrics, c.NewGauge("client.in_flight", client.InFlightCount, clientTags))
				metrics = append(metrics, c.NewGauge("client.messages", client.MessageCount, clientTags))
				metrics = append(metrics, c.NewGauge("client.finished", client.FinishCount, clientTags))
				metrics = append(metrics, c.NewGauge("client.requeued", client.RequeueCount, clientTags))
			}
		}
	}

	result := []Metric{}
	for i := range metrics {
		if metrics[i].Name != "" {
			result = append(result, metrics[i])
		}
	}

	log.WithFields(log.Fields{"node": c.Producer.Hostname}).Infof(`collected metrics for node %s`, c.Producer.Hostname)

	return result, nil
}
