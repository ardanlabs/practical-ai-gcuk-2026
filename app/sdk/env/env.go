// Package env provides support for reading configuration from the environment
// with fallback defaults.
package env

import (
	"os"
	"strconv"
	"time"
)

// String returns the value for the key or the default when not set.
func String(key string, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	return def
}

// Int returns the value for the key or the default when not set or not an int.
func Int(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}

	return def
}

// Bool returns the value for the key or the default when not set or not a bool.
func Bool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}

	return def
}

// Duration returns the value for the key or the default when not set or not a
// duration.
func Duration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}

	return def
}
