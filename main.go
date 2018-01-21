package main

import (
	"bufio"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/allegro/bigcache"
	"github.com/andrewl/wherearetheyflyingto/airport"
	"github.com/andrewl/wherearetheyflyingto/destinationfinder"
	"github.com/andrewl/wherearetheyflyingto/sbsmessage"
	"github.com/go-kit/kit/log"
	_ "github.com/mattn/go-sqlite3"
)

// Logger for logging
var logger log.Logger

// For cacheing the flights we've seen recently in order to combine multiple messages for a single flight
var flightcache *bigcache.BigCache

// sqlite3 db connection
var db *sql.DB

var destinations_cache destinationfinder.DestinationFinderCache

// lat long bounds of flights that we're interested in recording. A well setup aerial can pick up
// ads-b messages from hundreds of KMs away, but we want the ones just flying overhead.
var pos_lat float64
var pos_lon float64

// Use flightaware as our destination source
//var destination_finder destinationfinder.FlightAwareDestinationFinder
var destination_finder = destinationfinder.GetDestinationFinder()

func main() {

	// Initialise logging
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	logger.Log("msg", "Starting wherearetheyflyingto")

	// Open database and create table if necessary
	var err error
	db, err = sql.Open("sqlite3", "wherearetheyflyingto.db")
	if err != nil {
		logger.Log("msg", "Failed to create database", "err", err)
		return
	}
	defer db.Close()

	//@todo check if the table exists first? Perhaps doesn't matter?
	create_table_sql := `
	    create table watft (
			  destination_lat_long text,
			  destination_airport_name text,
				time datetime default current_timestamp,
				callsign text,
				altitude integer );
			`

	_, err = db.Exec(create_table_sql)
	if err != nil {
		logger.Log("msg", "Failed to create table. This probably isn't a problem if the table already exists", "err", err)
	}

	var write_heatmap = flag.Bool("heatmap", false, "just write out the heatmap and exit")
	flag.Parse()

	if *write_heatmap {
		heatmap_json := "["
		first := true
		for max_alt := 9000; max_alt < 45000; max_alt += 9000 {
			min_alt := max_alt - 9000
			rows, err := db.Query("select destination_lat_long, count(*) from watft where abs(altitude) > " + strconv.Itoa(min_alt) + " and abs(altitude) <= " + strconv.Itoa(max_alt) + " group by destination_lat_long")
			if err != nil {
				logger.Log("msg", "Failed to query database to create heatmap", "err", err)
				return
			}

			for rows.Next() {
				var destination_lat_long string
				var count string
				_ = rows.Scan(&destination_lat_long, &count)
				if first == false {
					heatmap_json = heatmap_json + ",\n"
				}
				heatmap_json = heatmap_json + "[" + destination_lat_long + "," + count + "," + strconv.Itoa(max_alt) + "]"
				first = false
			}
		}
		heatmap_json = heatmap_json + "]"

		file, err := os.Create("./watft_destinations.json")
		if err != nil {
			logger.Log("msg", "Failed to write watft_destinations.js", "err", err)
			return
		}
		defer file.Close()

		_, err = file.WriteString(heatmap_json)

		if err != nil {
			logger.Log("msg", "Failed to write to watf_destinations.js", "err", err)
			return
		}

		return
	}

	//open our cache
	destinations_cache.Open(db)

	pos_lat, _ = strconv.ParseFloat(os.Getenv("WATFT_LAT"), 64)
	pos_lon, _ = strconv.ParseFloat(os.Getenv("WATFT_LON"), 64)

	logger.Log("msg", "point set", "point", fmt.Sprintf("%v,%v", pos_lat, pos_lon))

	// Initialise our in-memory cache
	flightcache, _ = bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))

	server_ip := os.Getenv("WATFT_SERVER")
	// connect to this socket
	conn, err := net.Dial("tcp", server_ip)
	if err != nil {
		logger.Log("msg", "Failed to connect to server", "err", err)
		return
	}
	reader := bufio.NewReader(conn)
	for {
		// listen for messages
		message, _ := reader.ReadString('\n')
		if err != nil {
			logger.Log("msg", "Failed to read message from server", "err", err)
		}
		if err == nil {
			process_basestation_message(message)
		}
	}
}

func process_basestation_message(message string) {

	//logger.Log("message", message)

	var sbs_message sbsmessage.SBSMessage
	err := (&sbs_message).FromString(message)
	if err != nil {
		logger.Log("msg", "Failed to decode message", "err", err)
		return
	}

	// The flight id - if we don't have this we cannot correlate any of the messages.
	flightid, err := sbs_message.GetFlightId()

	if err != nil {
		logger.Log("msg", "Failed to find flight id")
		return
	}

	flight_seen, _ := flightcache.Get(flightid + "_seen")
	if flight_seen != nil {
		//we've already processed this flight and written it to permanent storage
		//so we can ignore it. The temporary cache will remove it.
		return
	}

	msg_callsign, err := sbs_message.GetCallsign()

	if err != nil {
		logger.Log("msg", "Failed to get callsign", "err", err)
	}

	if msg_callsign != "" {
		flightcache.Set(flightid+"_callsign", []byte(msg_callsign))
	}

	flight_has_been_overhead, _ := flightcache.Get(flightid + "_has_been_overhead")

	// If the flight hasn't been overhead then check to see if it is now
	if flight_has_been_overhead == nil {

		visible, err := is_visible_from(sbs_message, pos_lat, pos_lon)

		if err != nil {
			logger.Log("msg", "Failed to get whether visible", "err", err)
		}

		if visible {

			msg_alt, err := sbs_message.GetAltitude()

			if err != nil {
				logger.Log("msg", "Failed to get whether visible", "err", err)
			}
			if msg_alt != 0 {
				flightcache.Set(flightid+"_alt", []byte(strconv.Itoa(msg_alt)))
			}
			flightcache.Set(flightid+"_has_been_overhead", []byte("1"))
			logger.Log("msg", "Flight deemed to be overhead", "sbs", message)
		}
	}

	// Determine whether we have received all the information for this flight, and if so
	// attempt to determine the destination and write it to the db.
	// Might need to mutex this in order to prevent multiple writes?
	flight_callsign, _ := flightcache.Get(flightid + "_callsign")
	flight_alt, _ := flightcache.Get(flightid + "_alt")

	// Ensure we have a callsign, an altitude and the flight has actually been overhead before
	// proceeding any further.
	if flight_callsign == nil || flight_alt == nil || flight_has_been_overhead == nil {
		return
	}

	cached_dest_airport_code, _ := flightcache.Get(flightid + "_dest_airport_code")

	var dest_airport_code string

	if cached_dest_airport_code != nil {
		dest_airport_code = string(cached_dest_airport_code)
	}
	if dest_airport_code == "" {
		dest_airport_code, err = destination_finder.GetDestinationFromCallsign(string(flight_callsign))
		logger.Log("msg", "Got ", dest_airport_code)
		if err != nil {
			logger.Log("msg", "There was an error retrieving the destination from the callsign", "callsign", flight_callsign, "err", err)
			dest_airport_code = "error"
		} else {
			destinations_cache.Cache_set(string(flight_callsign), dest_airport_code)
		}
		flightcache.Set(flightid+"_dest_lat_long", []byte(dest_airport_code))
	}

	if dest_airport_code != "" && dest_airport_code != "error" {
		dest_airport, err := airport.GetAirportFromCode(dest_airport_code)
		if err != nil {
			logger.Log("msg", "Failed to get airport from airport code "+dest_airport_code)
		} else {
			dest_lat_long := fmt.Sprintf("%.4f,%.4f", dest_airport.Lat, dest_airport.Lon)
			_, err := db.Exec("insert into watft(destination_lat_long,destination_airport_name,callsign,altitude) values(?,?,?)", dest_lat_long, dest_airport.Name, flight_callsign, flight_alt)
			if err != nil {
				logger.Log("msg", string(flight_callsign)+" just flew overhead, but failed to write into db", "err", err)
			} else {
				logger.Log("msg", "A flight just passed overhead", "flight", string(flight_callsign), "altitute", flight_alt, "destination", dest_lat_long)
				flightcache.Set(flightid+"_seen", []byte("seen"))
			}
		}
	}
}

func is_visible_from(message sbsmessage.SBSMessage, pos_lat float64, pos_lon float64) (visibility bool, err error) {
	msg_lat, msg_lon, err := message.GetLatLong()

	if err != nil {
		return false, errors.New("Could not determine lat long from message")
	}

	/**
	alt, err := message.GetAltitude()

	if err != nil {
		return false, errors.New("Could not determine altitude from message")
	}
	//@todo calculate this based on altitude. the further up the plane is the more visible it's from
	*/

	var extents float64 = 0.02

	if msg_lat > (pos_lat-extents) && msg_lat < (pos_lat+extents) && msg_lon > (pos_lon-extents) && msg_lon < (pos_lon+extents) {
		return true, nil
	}

	return false, nil
}
