package destinationfinder

import "testing"
import "io/ioutil"
import "fmt"

func TestFlightAwareExtraction(t *testing.T) {
	var df FlightAwareDestinationFinder
	data, err := ioutil.ReadFile("./testdata/flightaware.json")
	if err != nil {
		fmt.Println(err)
		t.Error("Failed to open flightaware.json")
		return
	}
	airport_code, err := df.ExtractDestinationFromJson(data)
	if err != nil {
		fmt.Println("ExtractDestinationFromJSON errored with")
		fmt.Println(err)
		t.Fail()
	} else if airport_code != "EGGW" {
		fmt.Println("Failed to extract correct airport code: " + airport_code)
		t.Fail()
	}
}
