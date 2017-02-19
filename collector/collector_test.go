package collector

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fetcherMock struct{}

func (f fetcherMock) Fetch(url string) ([]byte, error) {
	return []byte(`{"status_code": 200}`), nil
}

func (f fetcherMock) GetURL(path string) string {
	panic("not implemented")
}

func (f fetcherMock) SetBaseURL(address string) {
	panic("not implemented")
}

type fetcherErrorMock struct{}

func (f fetcherErrorMock) Fetch(url string) ([]byte, error) {
	return []byte{}, fmt.Errorf("%s error", url)
}

func (f fetcherErrorMock) GetURL(path string) string {
	panic("not implemented")
}

func (f fetcherErrorMock) SetBaseURL(address string) {
	panic("not implemented")
}

func TestGetInfo_fetchError(t *testing.T) {
	collector := NSQDCollector{fetcher: fetcherErrorMock{}}
	_, err := collector.GetInfo()

	assert.Error(t, err)
}

func TestGetInfo(t *testing.T) {
	collector := NSQDCollector{fetcher: fetcherMock{}}
	info, err := collector.GetInfo()

	assert.NoError(t, err)
	assert.NotNil(t, info)
}

type fetcherInvalidStatusCodeErrorMock struct{}

func (f fetcherInvalidStatusCodeErrorMock) Fetch(url string) ([]byte, error) {
	return []byte(`{"status_code": 500}`), nil
}

func (f fetcherInvalidStatusCodeErrorMock) GetURL(path string) string {
	panic("not implemented")
}

func (f fetcherInvalidStatusCodeErrorMock) SetBaseURL(address string) {
	panic("not implemented")
}

func TestGetInfo_invalidStatusCode(t *testing.T) {
	collector := NSQDCollector{fetcher: fetcherInvalidStatusCodeErrorMock{}}
	_, err := collector.GetInfo()

	assert.EqualError(t, err, "response code was 500")
}

type fetcherInvalidJSONErrorMock struct{}

func (f fetcherInvalidJSONErrorMock) Fetch(url string) ([]byte, error) {
	return []byte("foo"), nil
}

func (f fetcherInvalidJSONErrorMock) GetURL(path string) string {
	panic("not implemented")
}

func (f fetcherInvalidJSONErrorMock) SetBaseURL(address string) {
	panic("not implemented")
}

func TestGetNodes_fetchJSONError(t *testing.T) {
	collector := NSQDCollector{fetcher: fetcherInvalidJSONErrorMock{}}
	_, err := collector.GetNodes()

	assert.EqualError(t, err, "invalid character 'o' in literal false (expecting 'a')")
}

func TestGetNodes_fetchError(t *testing.T) {
	collector := NSQDCollector{fetcher: fetcherErrorMock{}}
	_, err := collector.GetNodes()

	assert.Error(t, err)
}

func TestGetNodes(t *testing.T) {
	collector := NSQDCollector{fetcher: fetcherMock{}}
	nodes, err := collector.GetNodes()

	assert.NoError(t, err)
	assert.NotNil(t, nodes)
}
