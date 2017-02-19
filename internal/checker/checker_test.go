package checker_test

import (
	"errors"
	"testing"

	"github.com/ruimarinho/nsq-dogstatsd/internal"
	"github.com/stretchr/testify/assert"
)

func TestCheckAddresses(t *testing.T) {
	tests := []struct {
		input    []string
		expected error
	}{
		{[]string{"http://foobar.com", "http://127.0.0.1", "https://127.0.0.1"}, errors.New("all invalid")},
		{[]string{"foo", "http://127.0.0.1"}, errors.New("some invalid")},
		{[]string{"foo"}, nil},
	}

	for _, test := range tests {
		result := internal.CheckAddresses(test.input)

		if test.expected != nil {
			assert.Error(t, result)
		} else {
			assert.Nil(t, result)
		}
	}
}
