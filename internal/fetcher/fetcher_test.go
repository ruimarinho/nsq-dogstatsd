package fetcher_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetcher_GetURL(t *testing.T) {
	fetcher := &NSQDFetcher{baseURL: "http://127.0.0.1:4151"}

	assert.Equal(t, fetcher.GetURL(""), "http://127.0.0.1:4151/")
}

func TestFetcher_SetBaseURL(t *testing.T) {
	fetcher := &NSQDFetcher{}
	fetcher.SetBaseURL("127.0.0.1:4151")

	assert.Equal(t, fetcher.baseURL, "http://127.0.0.1:4151")
}

func TestFetcher_Fetch_invalidPath(t *testing.T) {
	fetcher := newFetcher("127.0.0.1:4151")
	_, err := fetcher.Fetch("%")

	assert.Error(t, err)
}

func TestFetcher_Fetch_invalidStatusCode(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

	defer server.Close()

	nsqdURL, parseErr := url.Parse(server.URL)
	assert.NoError(t, parseErr)

	fetcher := newFetcher(nsqdURL.Host)
	_, err := fetcher.Fetch("")

	assert.EqualError(t, err, "response code was 500")
}

func TestFetcher_Fetch(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status_code": 200}`))
		}))

	defer server.Close()

	nsqdURL, parseErr := url.Parse(server.URL)
	assert.NoError(t, parseErr)

	fetcher := newFetcher(nsqdURL.Host)
	body, err := fetcher.Fetch("")

	assert.NoError(t, err)
	assert.Equal(t, string(body), `{"status_code": 200}`)
}
