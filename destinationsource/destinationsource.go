// Package destinationsource provides functionality for
// determining where a flight with a given callsign is
// destined.
package destinationsource

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

/**
 * Retrieves the lat long of the destination (as a simple string, we're not interested in doing
 * any real processing with this, just using it as an index. Uses the flightaware website as
 * a datasource, and parses some js embedded in the page. As such this is potentially
 * brittle, but the function defintion should stand, even if we were to plugin a different
 * data source.
 **/
type DestinationFinder interface {
	GetDestinationFromCallsign(callsign string) (lat_long string, err error)
}

type FlightAwareDestinationSource struct {
}

/**
 * Retrieves the lat long of the destination (as a simple string, we're not interested in doing
 * any real processing with this, just using it as an index. Uses the flightaware website as
 * a datasource, and parses some js embedded in the page. As such this is potentially
 * brittle, but the function defintion should stand, even if we were to plugin a different
 * data source.
 **/
func (destination_source *FlightAwareDestinationSource) GetDestinationFromCallsign(callsign string) (lat_long string, err error) {
	return "", nil
	flight_url := "http://flightaware.com/live/flight/" + callsign

	resp, err := http.Get(flight_url)
	if err != nil {
		return "", errors.New("Failed to retrieve flight details from " + flight_url)
	}
	defer resp.Body.Close()

	flightaware_html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("Failed to retrieve flight details from " + flight_url)
	}

	if strings.Index(string(flightaware_html), "destinationPoint") == -1 {
		return "", errors.New("Failed to destinationPoint in html " + flight_url)
	}

	tmp_strings := strings.Split(string(flightaware_html), "destinationPoint\":[")
	lat_long = strings.Split(tmp_strings[1], "]")[0]

	return lat_long, nil
}
