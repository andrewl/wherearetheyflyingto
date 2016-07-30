package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/allegro/bigcache"
	"github.com/go-kit/kit/log"
	_ "github.com/mattn/go-sqlite3"
)

// Logger for logging
var logger log.Logger

// For cacheing the flights we've seen recently in order to combine multiple messages for a single flight
var flightcache *bigcache.BigCache

// sqlite3 db connection
var db *sql.DB

func main() {

	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	logger.Log("msg", "Starting wherearetheyflyingto")

	var err error
	db, err = sql.Open("sqlite3", "wherearetheyflyingto.db")
	if err != nil {
		logger.Log("msg", "Failed to open database")
		return
	}
	defer db.Close()

	//@todo check if the table exists first? Perhaps doesn't matter?
	create_table_sql := `
	    create table watft (
			  destination_lat_long text,
				time datetime default current_timestamp,
				callsign text,
				altitude integer );
			`

	_, err = db.Exec(create_table_sql)
	if err != nil {
		logger.Log("msg", "create table failed. This probably isn't a problem if the table already exists.", "err", err)
	}

	var write_heatmap = flag.Bool("heatmap", false, "just write out the heatmap and exit")
	flag.Parse()

	if *write_heatmap {
		rows, err := db.Query("select destination_lat_long, count(*) from watft group by destination_lat_long")
		if err != nil {
			logger.Log("msg", "Failed to query database to create heatmap", "err", err)
			return
		}

		heatmap_json := "var destination_points = [\n"
		for rows.Next() {
			var destination_lat_long string
			var count string
			_ = rows.Scan(&destination_lat_long, &count)
			heatmap_json = heatmap_json + "[" + destination_lat_long + "," + count + "]\n"
		}
		heatmap_json = heatmap_json + "]"

		file, err := os.Create("js/destinations.js")
		if err != nil {
			logger.Log("msg", "Failed to write destinations.js", "err", err)
			return
		}
		defer file.Close()

		w := bufio.NewWriter(file)
		fmt.Fprint(w, heatmap_json)
		return
	}

	flightcache, _ = bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))

	server_ip := os.Getenv("WATFT_SERVER")
	// connect to this socket
	conn, err := net.Dial("tcp", server_ip)
	if err != nil {
		logger.Log("msg", "Failed to connect to server", "err", err)
	}
	for {
		// listen for messages
		message, _ := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			logger.Log("msg", "Failed to read message from server", "err", err)
		}
		if err == nil {
			go process_basestation_message(message)
		}
	}
}

/**
Basestation messages consist of csv fields described here
@see http://woodair.net/SBS/Article/Barebones42_Socket_Data.htm
Different message types contain different data, eg callsign information
is not usually transmitted alongside positional information, however
a common 'flight id' is transmitted for each flight so different attributes
such as call sign, altitude, position etc can be tied together using the
flightid field. We use this field as the prefix of a key in our in-memory
cache and when we have all the information that we require in the cache then
we can write it out to a single file.
*/
func process_basestation_message(message string) {

	//Split the csv message into fields
	reader := csv.NewReader(strings.NewReader(message))
	message_record, err := reader.Read()

	if err != nil {
		logger.Log("msg", "Failed to decode message", "message", message)
		return
	}

	//Cursory validation that this is an SBS record
	//@todo - check the number of fields
	if message_record[0] != "MSG" {
		logger.Log("msg", "There following message does not appear to be an SBS message", "message", message)
		return
	}

	//The flight id
	flightid := message_record[4]

	flight_seen, _ := flightcache.Get(flightid + "_seen")
	if flight_seen != nil {
		//we've already processed this flight and written it to permanent storage
		//so we can ignore it. The temporary cache will remove it.
		return
	}

	//Any callsign or location information in this message. NB not all data is passed
	//with each message, but flightid is guaranteedd
	msg_callsign := message_record[10]
	msg_lat := message_record[14]
	msg_lon := message_record[15]

	if msg_callsign != "" {
		flightcache.Set(flightid+"_callsign", []byte(msg_callsign))
		dest_lat_long, _ := get_flight_destination_from_callsign(msg_callsign)
		if dest_lat_long != "" {
			flightcache.Set(flightid+"_dest_lat_long", []byte(dest_lat_long))
		}
	}

	if msg_lat != "" && msg_lon != "" {
		//@todo ignore if outside our bounds
		flightcache.Set(flightid+"_pos", []byte(msg_lon+","+msg_lat))
	}

	flight_pos, _ := flightcache.Get(flightid + "_pos")
	flight_callsign, _ := flightcache.Get(flightid + "_callsign")
	flight_dest_lat_long, _ := flightcache.Get(flightid + "_dest_lat_long")

	if flight_pos != nil && flight_dest_lat_long != nil && flight_callsign != nil {
		_, err := db.Exec("insert into watft(destination_lat_long,callsign,altitude) values(?,?,?)", flight_dest_lat_long, flight_callsign, 123)
		if err != nil {
			logger.Log("msg", "Failed to insert "+string(flight_callsign)+" into db", "err", err)
		} else {
			logger.Log("msg", "Written "+string(flight_callsign)+" to db")
			flightcache.Set(flightid+"_seen", []byte("seen"))
		}
	}
}

/**
 * Retrieves the lat long of the destination (as a simple string, we're not interested in doing
 * any real processing with this, just using it as an index. Uses the flightaware website as
 * a datasource, and parses some js embedded in the page. As such this is potentially
 * brittle, but the function defintion should stand, even if we were to plugin a different
 * data source.
 **/
func get_flight_destination_from_callsign(callsign string) (lat_long string, err error) {

	flight_url := "http://flightaware.com/live/flight/" + callsign

	resp, err := http.Get(flight_url)
	defer resp.Body.Close()
	if err != nil {
		logger.Log("msg", "Failed to retrieve flight details from flightaware", "flight_url", flight_url, "err", err)
		return "", errors.New("Failed to retrieve flight details")
	}

	flightaware_html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log("msg", "Failed to read flight details from flightaware", "flight_url", flight_url, "err", err)
		return "", errors.New("Failed to retrieve flight details")
	}

	if strings.Index(string(flightaware_html), "destinationPoint") == -1 {
		logger.Log("msg", "Failed to find destination in flight aware", "flight_url", flight_url)
		return "", errors.New("Failed to destination")
	}

	tmp_strings := strings.Split(string(flightaware_html), "destinationPoint\":[")
	lat_long = strings.Split(tmp_strings[1], "]")[0]

	return lat_long, nil
}
