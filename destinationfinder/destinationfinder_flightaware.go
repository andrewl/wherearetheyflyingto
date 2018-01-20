// Package destinationfinder provides functionality for
// determining where a flight with a given callsign is
// destined.
package destinationfinder

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type FlightAwareDestinationFinder struct {
}

/**
 * Retrieves the lat long of the destination (as a simple string, we're not interested in doing
 * any real processing with this, just using it as an index. Uses the flightaware website as
 * a datasource, and parses some js embedded in the page. As such this is potentially
 * brittle, but the function defintion should stand, even if we were to plugin a different
 * data source.
 **/
func (destination_finder *FlightAwareDestinationFinder) GetDestinationFromCallsign(callsign string) (lat_long string, err error) {
	flight_url := "http://" + os.Getenv("WATFT_FA_USERNAME") + ":" + os.Getenv("WATFT_FA_APIKEY") + "@flightxml.flightaware.com/json/FlightXML3/FlightInfoStatus?ident=" + callsign

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

	return airport_code, airport_name, lat_long, nil
}
