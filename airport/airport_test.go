package airport

import "testing"
import "fmt"

func TestAirport(t *testing.T) {
	airport, err := GetAirportFromCode("EGGW")
	if err != nil {
		fmt.Println("ExtractDestinationFromJSON errored with")
		fmt.Println(err)
		t.Fail()
	} else if airport.code != "EGGW" {
		fmt.Println("Failed to extract correct airport code: " + airport.code)
		t.Fail()
	}
}
