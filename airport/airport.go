// Package airport provides functionality for
// determining where a flight with a given callsign is
// destined.
package airport

import (
	"encoding/csv"
	"errors"
	"log"
	"os"
)

type AirportStruct struct {
	code string
	name string
	lat  string
	lon  string
}

func GetAirportFromCode(airport_code string) (airport AirportStruct, err error) {

	file, err := os.Open("./airports.dat")
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
		return airport, err
	}

	for _, v := range records {
		if v[5] == airport_code {
			airport.code = v[5]
			airport.name = v[1] + ", " + v[3]
			airport.lat = v[7]
			airport.lon = v[6]
			return airport, nil
		}
	}

	return airport, errors.New("Failed to find airport code " + airport_code)

}
