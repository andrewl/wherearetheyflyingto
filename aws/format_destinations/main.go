package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Destinations struct {
	ConsumedCapacity interface{} `json:"ConsumedCapacity"`
	Count            int         `json:"Count"`
	Items            []struct {
		Destination struct {
			S string `json:"S"`
		} `json:"destination"`
		Altitude struct {
			S string `json:"S"`
		} `json:"altitude"`
	} `json:"Items"`
	ScannedCount int `json:"ScannedCount"`
}

func main() {

	raw, err := ioutil.ReadFile("./destinations.json")

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var d Destinations
	json.Unmarshal(raw, &d)

	fmt.Println("[")
	for idx, destination := range d.Items {
		fmt.Printf("[%s,0,%s]", destination.Destination.S, destination.Altitude.S)
		if idx < d.Count-1 {
			fmt.Println(",")
		} else {
			fmt.Println("")
		}
	}
	fmt.Println("]")

}
