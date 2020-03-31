package parser

import (
	"regexp"
)

// Parse a slice of strings into their equivalent regexp.
func Parse(filters []string) ([]*regexp.Regexp, error) {
	var patterns []*regexp.Regexp

	for _, filter := range filters {
		regexp, err := regexp.Compile(filter)
		if err != nil {
			return nil, err
		}

		patterns = append(patterns, regexp)
	}

	return patterns, nil
}
