// Package destinationfinder provides functionality for
// determining where a flight with a given callsign is
// destined.
package destinationfinder

import (
	"os"
)

/**
 * Retrieves the airport code of the destination
 * (as a simple string, we're not interested in doing
 **/
type DestinationFinder interface {
	GetDestinationFromCallsign(callsign string) (airport_code string, err error)
}

func GetDestinationFinder() DestinationFinder {
	findername := os.Getenv("WATFT_FINDER")
	switch findername {
	default:
		return FlightAwareDestinationFinder{}
	}
	return nil
}
