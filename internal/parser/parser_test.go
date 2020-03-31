package parser_test

import (
	"regexp/syntax"
	"testing"

	. "github.com/ruimarinho/nsq-dogstatsd/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestParser_Parse_Invalid(t *testing.T) {
	patterns, err := Parse([]string{"*depth*"})

	assert.Nil(t, patterns)
	assert.Equal(t, err, &syntax.Error{Code: syntax.ErrMissingRepeatArgument, Expr: "*"})
}

func TestParser_Parse(t *testing.T) {
	patterns, err := Parse([]string{".*depth*."})

	assert.NotNil(t, patterns)
	assert.Nil(t, err)
}
