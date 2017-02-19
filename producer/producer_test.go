package producer_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProducer_GetTags(t *testing.T) {
	producer := Producer{Hostname: "localhost"}
	tags := producer.GetTags()

	assert.Equal(t, tags, []string{"node:localhost"})
}

func TestProducer_GetStats(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status_code": 200}`))
		}))

	defer server.Close()

	url, err := url.Parse(server.URL)
	assert.NoError(t, err)

	host, strPort, err := net.SplitHostPort(url.Host)
	assert.NoError(t, err)

	port, err := strconv.Atoi(strPort)
	assert.NoError(t, err)

	producer := Producer{BroadcastAddress: host, HTTPPort: port}
	stats, err := producer.GetStats()

	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestProducer_GetStats_error(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

	defer server.Close()

	url, parseErr := url.Parse(server.URL)
	assert.NoError(t, parseErr)

	host, strPort, splitErr := net.SplitHostPort(url.Host)
	assert.NoError(t, splitErr)

	port, convErr := strconv.Atoi(strPort)
	assert.NoError(t, convErr)

	producer := Producer{BroadcastAddress: host, HTTPPort: port}
	_, err := producer.GetStats()

	assert.Error(t, err)
}

func TestProducer_GetStats_unmarshalError(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status_code": 200`))
		}))

	defer server.Close()

	url, parseErr := url.Parse(server.URL)
	assert.NoError(t, parseErr)

	host, strPort, splitErr := net.SplitHostPort(url.Host)
	assert.NoError(t, splitErr)

	port, convErr := strconv.Atoi(strPort)
	assert.NoError(t, convErr)

	producer := Producer{BroadcastAddress: host, HTTPPort: port}
	_, err := producer.GetStats()

	assert.EqualError(t, err, "unexpected end of JSON input")
}
