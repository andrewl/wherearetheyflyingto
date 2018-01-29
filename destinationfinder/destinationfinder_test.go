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
		fmt.Println("Retrieved code " + airport_code + " rather than EGGW")
		t.Fail()
	}
}

func TestFlightAwareMalformedData(t *testing.T) {
	var df FlightAwareDestinationFinder
	data, err := ioutil.ReadFile("./testdata/malformed.json")
	if err != nil {
		fmt.Println(err)
		t.Error("Failed to open malformed.json")
		return
	}
	_, err = df.ExtractDestinationFromJson(data)
	if err == nil {
		fmt.Println("ExtractDestinationFromJSON should have errored but didn't.")
		t.Fail()
	}
}
