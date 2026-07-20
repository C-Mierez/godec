// Package timeutil provides time formatting helpers.
package timeutil

import "time"

// TimeNow returns the current UTC time formatted as RFC3339.
func TimeNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}
