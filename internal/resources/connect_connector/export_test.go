package connect_connector

import "time"

// SetPollInterval lets the tests wait for a fake connector without delay.
func SetPollInterval(d time.Duration) func() {
	previous := pollInterval
	pollInterval = d
	return func() { pollInterval = previous }
}
