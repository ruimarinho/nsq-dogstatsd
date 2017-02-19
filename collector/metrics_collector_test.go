package collector

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGauge(t *testing.T) {
	metric := newGauge("qux", float64(1), []string{"foo:tag"})

	assert.Equal(t, metric, Metric{
		Name:  "qux",
		Value: 1,
		Rate:  1,
		Type:  "gauge",
		Tags:  []string{"foo:tag"},
	})
}

func TestGauge(t *testing.T) {
	var tests = []struct{ value interface{} }{
		{true},
		{int(1)},
		{int32(1)},
		{int64(1)},
		{uint64(1)},
	}

	for _, tt := range tests {
		metric := gauge("qux", tt.value, []string{"foo:tag"})

		assert.Equal(t, metric, Metric{
			Name:  "qux",
			Value: 1,
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"foo:tag"},
		})
	}
}

func TestGauge_invalid(t *testing.T) {
	assert.Panics(t, func() {
		gauge("foo", float64(1), []string{})
	})
}

func TestCollectMetrics(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{
				"status_code": 200,
				"status_txt": "OK",
				"data": {
					"version": "1.0.0-compat",
					"health": "OK",
					"start_time": 1515289281,
					"topics": [{
						"topic_name": "foobar3000",
						"channels": [{
							"channel_name": "foo",
							"depth": 1,
							"backend_depth": 1,
							"in_flight_count": 1,
							"deferred_count": 1,
							"message_count": 1,
							"requeue_count": 1,
							"timeout_count": 1,
							"clients": [{
								"name": "foo",
								"client_id": "foo",
								"hostname": "foo",
								"version": "",
								"remote_address": "foo",
								"state": 0,
								"ready_count": 1,
								"in_flight_count": 2,
								"message_count": 3,
								"finish_count": 4,
								"requeue_count": 5,
								"user_agent": "foo"
							}],
							"paused": true
						}],
						"depth": 1,
						"backend_depth": 3,
						"message_count": 4,
						"paused": false
					}],
					"memory": {
						"heap_objects": 6263,
						"heap_idle_bytes": 892928,
						"heap_in_use_bytes": 1695744,
						"heap_released_bytes": 0,
						"gc_pause_usec_100": 0,
						"gc_pause_usec_99": 0,
						"gc_pause_usec_95": 0,
						"next_gc_bytes": 4473924,
						"gc_total_runs": 0
					}
				}
			}`))
		}))

	defer server.Close()

	url, err := url.Parse(server.URL)
	assert.Nil(t, err)

	host, strPort, err := net.SplitHostPort(url.Host)
	assert.Nil(t, err)

	port, err := strconv.Atoi(strPort)
	assert.Nil(t, err)

	metrics, err := collectMetrics(Producer{BroadcastAddress: host, HTTPPort: port, Hostname: "localhost"})
	assert.Nil(t, err)

	expected := []Metric{
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost"},
			Name:  "topic.count",
			Value: 1,
		},
		Metric{
			Name:  "memory.heap_objects",
			Rate:  1,
			Tags:  []string{"node:localhost"},
			Type:  "gauge",
			Value: 6263,
		},
		Metric{
			Name:  "memory.heap_idle_bytes",
			Rate:  1,
			Tags:  []string{"node:localhost"},
			Type:  "gauge",
			Value: 892928,
		},
		Metric{
			Name:  "memory.heap_in_use_bytes",
			Rate:  1,
			Tags:  []string{"node:localhost"},
			Type:  "gauge",
			Value: 1.695744e+06,
		},
		Metric{
			Name:  "memory.heap_released_bytes",
			Rate:  1,
			Tags:  []string{"node:localhost"},
			Type:  "gauge",
			Value: 0,
		},
		Metric{
			Name:  "memory.gc_pause_usec_100",
			Rate:  1,
			Tags:  []string{"node:localhost"},
			Type:  "gauge",
			Value: 0,
		},
		Metric{
			Name:  "memory.gc_pause_usec_99",
			Rate:  1,
			Tags:  []string{"node:localhost"},
			Type:  "gauge",
			Value: 0,
		},
		Metric{
			Name:  "memory.gc_pause_usec_95",
			Rate:  1,
			Tags:  []string{"node:localhost"},
			Type:  "gauge",
			Value: 0,
		},
		Metric{
			Name:  "memory.next_gc_bytes",
			Rate:  1,
			Tags:  []string{"node:localhost"},
			Type:  "gauge",
			Value: 4.473924e+06,
		},
		Metric{
			Name:  "memory.gc_runs",
			Rate:  1,
			Tags:  []string{"node:localhost"},
			Type:  "gauge",
			Value: 0,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000"},
			Name:  "topic.channels",
			Value: 1,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000"},
			Name:  "topic.depth",
			Value: 1,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000"},
			Name:  "topic.backend_depth",
			Value: 3,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000"},
			Name:  "topic.messages",
			Value: 4,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000"},
			Name:  "topic.paused",
			Value: 0,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo"},
			Name:  "channel.depth",
			Value: 1,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo"},
			Name:  "channel.backend_depth",
			Value: 1,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo"},
			Name:  "channel.in_flight",
			Value: 1,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo"},
			Name:  "channel.deferred",
			Value: 1,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo"},
			Name:  "channel.messages",
			Value: 1,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo"},
			Name:  "channel.requeued",
			Value: 1,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo"},
			Name:  "channel.timed_out",
			Value: 1,
		},
		Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo"},
			Name:  "channel.clients",
			Value: 1,
		}, Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo"},
			Name:  "channel.paused",
			Value: 1,
		}, Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo", "client_id:foo", "client_agent:foo", "client_hostname:foo", "client_address:foo"},
			Name:  "client.state",
			Value: 0,
		}, Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo", "client_id:foo", "client_agent:foo", "client_hostname:foo", "client_address:foo"},
			Name:  "client.ready_count",
			Value: 1,
		}, Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo", "client_id:foo", "client_agent:foo", "client_hostname:foo", "client_address:foo"},
			Name:  "client.in_flight",
			Value: 2,
		}, Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo", "client_id:foo", "client_agent:foo", "client_hostname:foo", "client_address:foo"},
			Name:  "client.messages",
			Value: 3,
		}, Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo", "client_id:foo", "client_agent:foo", "client_hostname:foo", "client_address:foo"},
			Name:  "client.finished",
			Value: 4,
		}, Metric{
			Rate:  1,
			Type:  "gauge",
			Tags:  []string{"node:localhost", "topic:foobar3000", "channel:foo", "client_id:foo", "client_agent:foo", "client_hostname:foo", "client_address:foo"},
			Name:  "client.requeued",
			Value: 5,
		},
	}

	assert.Equal(t, metrics, expected)
}

func TestCollectMetrics_error(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status_code":500}`))
		}))

	defer server.Close()

	url, err := url.Parse(server.URL)
	assert.Nil(t, err)

	host, port, err := net.SplitHostPort(url.Host)
	assert.Nil(t, err)

	sport, err := strconv.Atoi(port)
	assert.Nil(t, err)

	_, errMetrics := collectMetrics(Producer{BroadcastAddress: host, HTTPPort: sport, Hostname: "net"})
	assert.NotNil(t, errMetrics)
}
