package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStaggerOffset(t *testing.T) {
	interval := 10 * time.Second

	assert.Equal(t, time.Duration(0), staggerOffset(0, 1, interval))

	assert.Equal(t, time.Duration(0), staggerOffset(0, 4, interval))
	assert.Equal(t, 2500*time.Millisecond, staggerOffset(1, 4, interval))
	assert.Equal(t, 5*time.Second, staggerOffset(2, 4, interval))
	assert.Equal(t, 7500*time.Millisecond, staggerOffset(3, 4, interval))
}

func TestStaggerOffset_withinInterval(t *testing.T) {
	interval := 5 * time.Second

	for n := 1; n <= 100; n++ {
		for i := 0; i < n; i++ {
			assert.Less(t, staggerOffset(i, n, interval), interval)
		}
	}
}
