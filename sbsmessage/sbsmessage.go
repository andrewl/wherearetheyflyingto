package sbsmessage

import (
	"encoding/csv"
	"errors"
	"strconv"
	"strings"
)

/**
 * Basestation messages consist of comma separated fields described here
 * @see http://woodair.net/SBS/Article/Barebones42_Socket_Data.htm
 *
 * Different message types contain different data, eg callsign information
 * is not usually transmitted alongside positional information, however
 * a common 'flight id' is transmitted for each flight so different attributes
 * such as call sign, altitude, position etc can be tied together using the
 * flightid field.
 */
type SBSMessage struct {
	fields    []string
	flight_id string
}

func (message *SBSMessage) FromString(message_string string) (err error) {

	//Split the csv message into fields
	reader := csv.NewReader(strings.NewReader(message_string))
	message.fields, err = reader.Read()

	if err != nil {
		return errors.New("Failed to decode message")
	}

	//Cursory validation that this is an SBS record
	if message.fields[0] != "MSG" {
		return errors.New("Not an SBD message")
	}

	//Save the flight id
	message.flight_id = message.fields[5]

	return nil
}

func (message SBSMessage) GetFlightId() (flightid string, err error) {
	return message.flight_id, nil
}

func (message SBSMessage) GetCallsign() (callsign string, err error) {
	//@todo - check that fields[10] exists
	return strings.TrimSpace(message.fields[10]), nil
}

func (message SBSMessage) GetLatLong() (lat float64, lon float64, err error) {
	lat, _ = strconv.ParseFloat(message.fields[14], 64)
	lon, _ = strconv.ParseFloat(message.fields[15], 64)
	return lat, lon, nil
}

func (message SBSMessage) GetAltitude() (altitude int, err error) {
	altitude, _ = strconv.Atoi(message.fields[11])
	return altitude, nil
}
