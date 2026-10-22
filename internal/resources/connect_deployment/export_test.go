package connect_deployment

import "time"

// SetPollInterval lets the tests wait for a fake deployment without delay.
func SetPollInterval(d time.Duration) func() {
	previous := pollInterval
	pollInterval = d
	return func() { pollInterval = previous }
}
