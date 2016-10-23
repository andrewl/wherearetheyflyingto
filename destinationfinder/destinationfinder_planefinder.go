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

type PlaneFinderDestinationFinder struct {
}

/**
 * Retrieves the lat long of the destination (as a simple string, we're not interested in doing
 * any real processing with this, just using it as an index. Uses the flightaware website as
 * a datasource, and parses some js embedded in the page. As such this is potentially
 * brittle, but the function defintion should stand, even if we were to plugin a different
 * data source.
 **/
func (destination_finder PlaneFinderDestinationFinder) GetDestinationFromCallsign(callsign string) (lat_long string, err error) {
	if callsign == "" {
		return "", errors.New("Not going to get latlong from an empty callsign")
	}
	flight_url := "http://www.planefinder.net/data/flight/" + callsign

	resp, err := http.Get(flight_url)
	if err != nil {
		return "", errors.New("Failed to retrieve flight details from " + flight_url)
	}
	defer resp.Body.Close()

	planefinder_html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("Failed to retrieve flight details from " + flight_url)
	}

	return destination_finder.ExtractDestinationFromHTML(planefinder_html)
}

func (destination_finder *PlaneFinderDestinationFinder) ExtractDestinationFromHTML(html []byte) (lat_long string, err error) {
	if strings.Index(string(html), "/data/airport") == -1 {
		return "", errors.New("Failed to arrival airport in html ")
	}

	tmp_strings := strings.Split(string(html), "/data/airport/")
	airport_code := strings.Split(tmp_strings[2], "\"")[0]
	return destination_finder.getLatLongFromAirportCode(airport_code)
}

func (destination_finder *PlaneFinderDestinationFinder) getLatLongFromAirportCode(airport_code string) (lat_long string, err error) {
	if airport_code == "" {
		return "", errors.New("Not going to get latlong from an empty airport code")
	}

	airport_html, err := ioutil.ReadFile("./airport_" + airport_code + ".cache")
	if err != nil {
		airport_code_url := "http://www.planefinder.net/data/airport/" + airport_code

		resp, err := http.Get(airport_code_url)
		if err != nil {
			return "", errors.New("Failed to retrieve airport details from " + airport_code_url)
		}
		defer resp.Body.Close()

		airport_html, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", errors.New("Failed to retrieve airport details from " + airport_code_url)
		}
		cache_file, err := os.Create("./airport_" + airport_code + ".cache")
		if err == nil {
			cache_file.Write(airport_html)
			cache_file.Sync()
			cache_file.Close()
		}
	}

	if strings.Index(string(airport_html), "location=") == -1 {
		return "", errors.New("Failed to arrival airport in html ")
	}

	tmp_strings := strings.Split(string(airport_html), "location=")
	airport_lat_lng := strings.Split(tmp_strings[1], ",13")[0]

	return airport_lat_lng, nil
}
