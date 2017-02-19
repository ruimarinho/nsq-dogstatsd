package fetcher

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// Fetcher fetches the content of a URL.
type Fetcher interface {
	Fetch(url string) ([]byte, error)
	GetURL(path string) string
	SetBaseURL(address string)
}

// NSQDFetcher holds the baseURL to the nsqd node which includes the HTTP scheme.
type NSQDFetcher struct {
	baseURL string
}

// Fetch retrieves data from a remote resource.
func (f NSQDFetcher) Fetch(path string) ([]byte, error) {
	response, err := http.Get(f.GetURL(path))
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("response code was %d", response.StatusCode)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// GetURL returns the URL constructed by joining the base URL with the given path.
func (f NSQDFetcher) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", f.baseURL, path)
}

// SetBaseURL sets the base URL for the remote resource.
func (f *NSQDFetcher) SetBaseURL(address string) {
	f.baseURL = fmt.Sprintf("http://%s", address)
}

// NewFetcher instantiates a NSQDFetcher and sets its base URL.
func NewFetcher(address string) Fetcher {
	fetcher := &NSQDFetcher{}
	fetcher.SetBaseURL(address)

	return fetcher
}
