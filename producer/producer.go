package producer

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/ruimarinho/nsq-dogstatsd/internal/fetcher"
)

// Producer represents a nsqd node.
type Producer struct {
	Version          string `json:"version"`
	RemoteAddress    string `json:"remote_address,omitempty"`
	BroadcastAddress string `json:"broadcast_address"`
	Hostname         string `json:"hostname"`
	HTTPPort         int    `json:"http_port"`
	TCPPort          int64  `json:"tcp_port"`
	StartTime        int    `json:"start_time,omitempty"`
}

// GetTags returns the Producer tags including, by default, a tag with its hostname.
func (p Producer) GetTags() []string {
	return []string{fmt.Sprintf("node:%s", p.Hostname)}
}

// Stats wraps /stats data.
type Stats struct {
	StatusCode int       `json:"status_code"`
	StatusTxt  string    `json:"status_txt"`
	Data       StatsData `json:"data"`
}

// MemoryStats wraps /stats memory data.
type MemoryStats struct {
	HeapObjects       uint64 `json:"heap_objects"`
	HeapIdleBytes     uint64 `json:"heap_idle_bytes"`
	HeapInUseBytes    uint64 `json:"heap_in_use_bytes"`
	HeapReleasedBytes uint64 `json:"heap_released_bytes"`
	GCPauseUsec100    uint64 `json:"gc_pause_usec_100"`
	GCPauseUsec99     uint64 `json:"gc_pause_usec_99"`
	GCPauseUsec95     uint64 `json:"gc_pause_usec_95"`
	NextGCBytes       uint64 `json:"next_gc_bytes"`
	GCTotalRuns       uint32 `json:"gc_total_runs"`
}

// LatencyStats wraps /stats end-to-end processing latency data.
type LatencyStats struct {
	Count       int                  `json:"count"`
	Percentiles []map[string]float64 `json:"percentiles"`
}

// ClientStats wraps /stats client data.
type ClientStats struct {
	ClientID      string `json:"client_id"`
	Hostname      string `json:"hostname"`
	Version       string `json:"version"`
	RemoteAddress string `json:"remote_address"`
	State         int32  `json:"state"`
	ReadyCount    int64  `json:"ready_count"`
	InFlightCount int64  `json:"in_flight_count"`
	MessageCount  uint64 `json:"message_count"`
	FinishCount   uint64 `json:"finish_count"`
	RequeueCount  uint64 `json:"requeue_count"`
	UserAgent     string `json:"user_agent"`
}

// ChannelStats wraps /stats channel data.
type ChannelStats struct {
	ChannelName          string        `json:"channel_name"`
	Depth                int64         `json:"depth"`
	BackendDepth         int64         `json:"backend_depth"`
	InFlightCount        int           `json:"in_flight_count"`
	DeferredCount        int           `json:"deferred_count"`
	MessageCount         uint64        `json:"message_count"`
	RequeueCount         uint64        `json:"requeue_count"`
	TimeoutCount         uint64        `json:"timeout_count"`
	Clients              []ClientStats `json:"clients"`
	Paused               bool          `json:"paused"`
	E2eProcessingLatency *LatencyStats `json:"e2e_processing_latency"`
}

// TopicStats wraps /stats topic data.
type TopicStats struct {
	TopicName            string         `json:"topic_name"`
	Channels             []ChannelStats `json:"channels"`
	Depth                int64          `json:"depth"`
	BackendDepth         int64          `json:"backend_depth"`
	MessageCount         uint64         `json:"message_count"`
	Paused               bool           `json:"paused"`
	E2eProcessingLatency *LatencyStats  `json:"e2e_processing_latency"`
}

// StatsData is an embedded Stats type.
type StatsData struct {
	Version   string       `json:"version"`
	Health    string       `json:"health"`
	StartTime int64        `json:"start_time"`
	Topics    []TopicStats `json:"topics"`
	Memory    MemoryStats  `json:"memory"`
}

// GetStats retrieves and parses the statistics of a nsqd.
func (p Producer) GetStats() (Stats, error) {
	var stats Stats

	fetcher := fetcher.NewFetcher(p.HTTPAddress())
	body, err := fetcher.Fetch("stats?format=json")
	if err != nil {
		return stats, err
	}

	err = json.Unmarshal(body, &stats)
	if err != nil {
		return stats, err
	}

	if stats.StatusCode != 200 {
		return stats, fmt.Errorf("response code was %d", stats.StatusCode)
	}

	return stats, err
}

// HTTPAddress returns the broadcast address (e.g. 127.0.0.1) joined together
// with the port (e.g 4151).
func (p Producer) HTTPAddress() string {
	return net.JoinHostPort(p.BroadcastAddress, strconv.Itoa(p.HTTPPort))
}
