package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type Destinations struct {
	ConsumedCapacity interface{} `json:"ConsumedCapacity"`
	Count            int         `json:"Count"`
	LambdaRow        []struct {
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

	var destination_count = make(map[string]int)
	var altitude_sum = make(map[string]int)

	for _, destination := range d.LambdaRow {
		if destination_count[destination.Destination.S] == 0 {
			destination_count[destination.Destination.S] = 1
			altitude_sum[destination.Destination.S], _ = strconv.Atoi(destination.Altitude.S)
		} else {
			destination_count[destination.Destination.S]++
			alt, _ := strconv.Atoi(destination.Altitude.S)
			altitude_sum[destination.Destination.S] += alt
		}
	}

	fmt.Println("[")
	var count = 0
	var total = len(destination_count)
	for k, v := range destination_count {
		fmt.Printf("[%s,%d,%d]", k, v, altitude_sum[k]/v)
		if count < total-1 {
			fmt.Println(",")
		} else {
			fmt.Println("")
		}
		count++
	}
	fmt.Println("]")

}
