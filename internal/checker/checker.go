package checker

import (
	"errors"
	"strings"
)

// CheckAddresses checks if an address incorrectly includes the HTTP scheme.
func CheckAddresses(addresses []string) error {
	for _, address := range addresses {
		if strings.HasPrefix(address, "http") {
			return errors.New("address should not contain http scheme")
		}
	}

	return nil
}
