//  Package destinationfinder provides functionality for
// determining where a flight with a given callsign is
// destined.
package destinationfinder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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
func (destination_finder FlightAwareDestinationFinder) GetDestinationFromCallsign(callsign string) (airport_code string, err error) {

	flight_url := "http://" + os.Getenv("WATFT_FA_USERNAME") + ":" + os.Getenv("WATFT_FA_APIKEY") + "@flightxml.flightaware.com/json/FlightXML3/FlightInfoStatus?ident=" + callsign

	resp, err := http.Get(flight_url)
	if err != nil {
		return "", errors.New("Failed to retrieve flight details from " + flight_url)
	}
	defer resp.Body.Close()

	flightaware_json, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", errors.New("Failed to retrieve flight details from " + flight_url)
	}

	return destination_finder.ExtractDestinationFromJson(flightaware_json)
}

func (destination_finder *FlightAwareDestinationFinder) ExtractDestinationFromJson(json_string []byte) (airport_code string, err error) {

	var js json.RawMessage
	if json.Unmarshal(json_string, &js) != nil {
		return "", errors.New(fmt.Sprintf("Invalid JSON: %s", string(json_string)))
	}

	type FlightAwareResult struct {
		FlightInfoStatusResult struct {
			NextOffset int `json:"next_offset"`
			Flights    []struct {
				Origin struct {
					AlternateIdent string `json:"alternate_ident"`
					City           string `json:"city"`
					Code           string `json:"code"`
					AirportName    string `json:"airport_name"`
				} `json:"origin"`
				EstimatedArrivalTime struct {
					Time      string `json:"time"`
					Tz        string `json:"tz"`
					Dow       string `json:"dow"`
					Epoch     int    `json:"epoch"`
					Date      string `json:"date"`
					Localtime int    `json:"localtime"`
				} `json:"estimated_arrival_time"`
				Adhoc        bool   `json:"adhoc"`
				Flightnumber string `json:"flightnumber"`
				Destination  struct {
					AirportName    string `json:"airport_name"`
					Code           string `json:"code"`
					City           string `json:"city"`
					AlternateIdent string `json:"alternate_ident"`
				} `json:"destination"`
				DepartureDelay    int    `json:"departure_delay"`
				DistanceFiled     int    `json:"distance_filed"`
				Cancelled         bool   `json:"cancelled"`
				Type              string `json:"type"`
				ActualArrivalTime struct {
					Time      string `json:"time"`
					Dow       string `json:"dow"`
					Tz        string `json:"tz"`
					Epoch     int    `json:"epoch"`
					Date      string `json:"date"`
					Localtime int    `json:"localtime"`
				} `json:"actual_arrival_time"`
				FaFlightID          string `json:"faFlightID"`
				Airline             string `json:"airline"`
				Ident               string `json:"ident"`
				ActualDepartureTime struct {
					Date      string `json:"date"`
					Localtime int    `json:"localtime"`
					Time      string `json:"time"`
					Dow       string `json:"dow"`
					Tz        string `json:"tz"`
					Epoch     int    `json:"epoch"`
				} `json:"actual_departure_time"`
				ProgressPercent    int `json:"progress_percent"`
				FiledDepartureTime struct {
					Localtime int    `json:"localtime"`
					Date      string `json:"date"`
					Time      string `json:"time"`
					Tz        string `json:"tz"`
					Dow       string `json:"dow"`
					Epoch     int    `json:"epoch"`
				} `json:"filed_departure_time"`
				Status                 string `json:"status"`
				Diverted               bool   `json:"diverted"`
				Blocked                bool   `json:"blocked"`
				EstimatedDepartureTime struct {
					Date      string `json:"date"`
					Localtime int    `json:"localtime"`
					Time      string `json:"time"`
					Tz        string `json:"tz"`
					Dow       string `json:"dow"`
					Epoch     int    `json:"epoch"`
				} `json:"estimated_departure_time"`
				Tailnumber       string `json:"tailnumber"`
				FiledArrivalTime struct {
					Epoch int `json:"epoch"`
				} `json:"filed_arrival_time"`
			} `json:"flights"`
		} `json:"FlightInfoStatusResult"`
	}

	var flightAwareResult FlightAwareResult

	err = json.Unmarshal(json_string, &flightAwareResult)

	if err == nil {
		if len(flightAwareResult.FlightInfoStatusResult.Flights) > 0 {
			airport_code = flightAwareResult.FlightInfoStatusResult.Flights[0].Destination.Code
		} else {
			err = errors.New(fmt.Sprintf("Error no flights in payload:\nJSON Payload: %s", err, string(json_string)))
		}
	} else {
		err = errors.New(fmt.Sprintf("Error: %v.\nJSON Payload: %s", err, string(json_string)))
	}

	return airport_code, err
}
