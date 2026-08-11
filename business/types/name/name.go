// Package name represents a name in the system.
package name

import (
	"errors"
	"fmt"
	"regexp"
)

var nameRegEx = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_ -]{2,20}$")

// ErrInvalidName is returned when the name is not valid.
var ErrInvalidName = errors.New("invalid name")

// Name represents a name in the system.
type Name struct {
	value string
}

// String returns the value of the name.
func (n Name) String() string {
	return n.value
}

// Equal provides support for the go-cmp package and testing.
func (n Name) Equal(n2 Name) bool {
	return n.value == n2.value
}

// MarshalText provides support for logging and any marshal needs.
func (n Name) MarshalText() ([]byte, error) {
	return []byte(n.value), nil
}

// Parse parses the string value and returns a name if the value complies
// with the rules for a name.
func Parse(value string) (Name, error) {
	if !nameRegEx.MatchString(value) {
		return Name{}, ErrInvalidName
	}

	return Name{value}, nil
}

// MustParse parses the string value and returns a name if the value
// complies with the rules for a name. If an error occurs the function panics.
func MustParse(value string) Name {
	name, err := Parse(value)
	if err != nil {
		panic(fmt.Sprintf("parsing name %q: %s", value, err))
	}

	return name
}
