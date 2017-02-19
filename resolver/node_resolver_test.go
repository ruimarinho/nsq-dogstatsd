package resolver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveNodes_NSQLookupdAddresses(t *testing.T) {
	nsqlookupdServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{
              "status_code": 200,
              "status_txt": "OK",
              "data": {
                "producers": [
                  {
                    "broadcast_address": "127.0.0.1",
                    "http_port": 4151
                  }
                ]
              }
            }`))
		}))

	defer nsqlookupdServer.Close()

	nsqlookupdURL, err := url.Parse(nsqlookupdServer.URL)
	assert.Nil(t, err)

	producers, err := resolveNodes([]string{}, []string{nsqlookupdURL.Host})
	assert.Nil(t, err)
	assert.Len(t, producers, 1)
}

func TestResolveNodes_NSQLookupdAddresses_Error(t *testing.T) {
	nsqlookupdServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status_code": 500}`))
		}))

	defer nsqlookupdServer.Close()

	nsqlookupdURL, parseErr := url.Parse(nsqlookupdServer.URL)
	assert.Nil(t, parseErr)

	_, err := resolveNodes([]string{}, []string{nsqlookupdURL.Host})
	assert.NotNil(t, err)
}

func TestResolveNodes_NSQDAddresses(t *testing.T) {
	nsqdServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{
              "status_code": 200,
              "status_txt": "OK",
              "data": {
                "broadcast_address": "127.0.0.1",
                "hostname": "58d493c00ddc",
                "http_port": 4151
              }
            }`))
		}))

	defer nsqdServer.Close()

	nsqdURL, err := url.Parse(nsqdServer.URL)
	assert.Nil(t, err)

	producers, err := resolveNodes([]string{nsqdURL.Host}, []string{})
	assert.Nil(t, err)
	assert.Len(t, producers, 1)
}

func TestResolveNodes_NSQDAddresses_Error(t *testing.T) {
	nsqdServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status_code": 500`))
		}))

	defer nsqdServer.Close()

	nsqdURL, parseErr := url.Parse(nsqdServer.URL)
	assert.Nil(t, parseErr)

	_, err := resolveNodes([]string{nsqdURL.Host, nsqdURL.Host}, []string{})
	assert.NotNil(t, err)
}

func TestResolveNodes_NSQDAddresses_Duplicates_Ignored(t *testing.T) {
	nsqdServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{
              "status_code": 200,
              "status_txt": "OK",
              "data": {
                "broadcast_address": "127.0.0.1",
                "hostname": "58d493c00ddc",
                "http_port": 4151
              }
            }`))
		}))

	defer nsqdServer.Close()

	nsqdURL, err := url.Parse(nsqdServer.URL)
	assert.Nil(t, err)

	producers, err := resolveNodes([]string{nsqdURL.Host, nsqdURL.Host}, []string{})
	assert.Nil(t, err)
	assert.Len(t, producers, 1)
}
