package slice

import "strings"

// StringSlice represents a slice of strings.
type StringSlice []string

// Set appends a new element to the string slice.
func (a *StringSlice) Set(s string) error {
	*a = append(*a, s)
	return nil
}

// String joins the elements of the string slice using commas.
func (a *StringSlice) String() string {
	return strings.Join(*a, ",")
}
