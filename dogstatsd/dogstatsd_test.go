package dogstatsd_test

import (
	"net"
	"testing"
	"time"

	. "github.com/ruimarinho/nsq-dogstatsd/dogstatsd"
	"github.com/stretchr/testify/assert"
)

func TestNewDogStatsdDClient(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	assert.Nil(t, err)
	defer conn.Close()

	client, err := NewDogStatsDClient(conn.LocalAddr().String(), "foobar", []string{"foo", "bar"})
	assert.Nil(t, err)
	defer client.Close()

	assert.Nil(t, client.Gauge("biz", 1, nil, 1))
	assert.Nil(t, client.Flush())

	buf := make([]byte, 1024)
	assert.Nil(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	n, _, err := conn.ReadFrom(buf)
	assert.Nil(t, err)
	assert.Equal(t, "foobar.biz:1|g|#foo,bar\n", string(buf[:n]))
}

func TestNewDogStatsdDClient_Invalid_Address(t *testing.T) {
	_, err := NewDogStatsDClient("foo", "foobar", []string{"foo", "bar"})

	assert.NotNil(t, err)
}
