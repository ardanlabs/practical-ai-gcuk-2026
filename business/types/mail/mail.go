// Package mail represents an email address in the system.
package mail

import (
	"fmt"
	"net/mail"
)

// Address represents an email address in the system.
type Address struct {
	value *mail.Address
}

// String returns the value of the address.
func (a Address) String() string {
	if a.value == nil {
		return ""
	}

	return a.value.Address
}

// Equal provides support for the go-cmp package and testing.
func (a Address) Equal(a2 Address) bool {
	return a.String() == a2.String()
}

// MarshalText provides support for logging and any marshal needs.
func (a Address) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// Parse parses the string value and returns an address if the value complies
// with the rules for an email address.
func Parse(value string) (Address, error) {
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return Address{}, err
	}

	return Address{addr}, nil
}

// MustParse parses the string value and returns an address if the value
// complies with the rules for an email address. If an error occurs the
// function panics.
func MustParse(value string) Address {
	addr, err := Parse(value)
	if err != nil {
		panic(fmt.Sprintf("parsing address %q: %s", value, err))
	}

	return addr
}
