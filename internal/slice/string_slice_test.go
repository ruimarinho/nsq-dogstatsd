package slice_test

import (
	"testing"

	. "github.com/ruimarinho/nsq-dogstatsd/internal/slice"
	"github.com/stretchr/testify/assert"
)

func TestStringsSlice_Set(t *testing.T) {
	s := StringSlice{"foo"}
	s.Set("bar")

	assert.EqualValues(t, s, []string{"foo", "bar"})
}

func TestStringsSlice_String(t *testing.T) {
	s := StringSlice{"foo", "bar"}

	assert.Equal(t, s.String(), "foo,bar")
}
