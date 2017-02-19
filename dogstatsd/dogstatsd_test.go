package dogstatsd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDogStatsdDClient(t *testing.T) {
	client, err := newDogStatsDClient("127.0.0.1:8125", "foobar", []string{"foo", "bar"})

	assert.Nil(t, err)
	assert.Equal(t, client.Namespace, "foobar.")
	assert.Equal(t, client.Tags, []string{"foo", "bar"})
}

func TestNewDogStatsdDClient_Invalid_Address(t *testing.T) {
	_, err := newDogStatsDClient("foo", "foobar", []string{"foo", "bar"})

	assert.NotNil(t, err)
}
