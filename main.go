package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"github.com/allegro/bigcache"
	"github.com/andrewl/wherearetheyflyingto/destinationsource"
	"github.com/andrewl/wherearetheyflyingto/sbsmessage"
	"github.com/go-kit/kit/log"
	_ "github.com/mattn/go-sqlite3"
	"net"
	"os"
	"strconv"
	"time"
)

// Logger for logging
var logger log.Logger

// For cacheing the flights we've seen recently in order to combine multiple messages for a single flight
var flightcache *bigcache.BigCache

// sqlite3 db connection
var db *sql.DB

// lat long bounds of flights that we're interested in recording. A well setup aerial can pick up
// ads-b messages from hundreds of KMs away, but we want the ones just flying overhead.
var min_lat float64
var min_lon float64
var max_lat float64
var max_lon float64

// Use flightaware as our destination source
var destination_source destinationsource.FlightAwareDestinationSource

func main() {

	// Initialise logging
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	logger.Log("msg", "Starting wherearetheyflyingto")

	// Open database and create table if necessary
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

	min_lat, _ = strconv.ParseFloat(os.Getenv("WATFT_MIN_LAT"), 64)
	min_lon, _ = strconv.ParseFloat(os.Getenv("WATFT_MIN_LON"), 64)
	max_lat, _ = strconv.ParseFloat(os.Getenv("WATFT_MAX_LAT"), 64)
	max_lon, _ = strconv.ParseFloat(os.Getenv("WATFT_MAX_LON"), 64)

	logger.Log("msg", "boundaries set", "boundaries", fmt.Sprintf("%v,%v -> %v,%v", min_lat, min_lon, max_lat, max_lon))

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

	var sbs_message sbsmessage.SBSMessage
	err := (&sbs_message).FromString(message)
	if err != nil {
		logger.Log("msg", "Failed to decode message", "err", err)
		return
	}

	//The flight id
	flightid, err := sbs_message.GetFlightId()

	if err != nil {
		logger.Log("msg", "Failed to find flight id")
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

	flight_destination_lat_long, err := flightcache.Get(flightid + "_dest_lat_long")
	if flight_destination_lat_long == nil {
		if msg_callsign != "" {
			flightcache.Set(flightid+"_callsign", []byte(msg_callsign))
			//@todo refactor this into an interface
			dest_lat_long, err := destination_source.GetDestinationFromCallsign(msg_callsign)
			if err != nil {
				logger.Log("msg", "There was an error retrieving the destination from the callsign", "err", err)
				flightcache.Set(flightid+"_dest_lat_long", []byte("error"))
			} else {
				flightcache.Set(flightid+"_dest_lat_long", []byte(dest_lat_long))
			}
		}
	} else {
		//We errored trying to get the lat-long so don't bother again
		//@todo - tighten up the error handling around this probably.
		//perhaps we do want to try again, but throttle it, so perhaps
		//write in the next time that we should try to retrieve the
		//information?
		if string(flight_destination_lat_long) == "error" {
			return
		}
	}

	flight_has_been_overhead, _ := flightcache.Get(flightid + "_has_been_overhead")

	if flight_has_been_overhead != nil {

		visible, err := is_visible_from(sbs_message, 0.0, 0.0)

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
		}
	}

	// Determine whether we have received all the information for this flight, and if so
	// write it to the db.
	// Might need to mutex this in order to prevent multiple writes?
	if flight_has_been_overhead != nil {
		return
	}

	flight_callsign, _ := flightcache.Get(flightid + "_callsign")
	flight_alt, _ := flightcache.Get(flightid + "_alt")
	flight_dest_lat_long, _ := flightcache.Get(flightid + "_dest_lat_long")

	if flight_dest_lat_long != nil && flight_callsign != nil && flight_alt != nil {
		_, err := db.Exec("insert into watft(destination_lat_long,callsign,altitude) values(?,?,?)", flight_dest_lat_long, flight_callsign, flight_alt)
		if err != nil {
			logger.Log("msg", string(flight_callsign)+" just flew overhead, but failed to write into db", "err", err)
		} else {
			logger.Log("msg", string(flight_callsign)+" just flew overhead writing to db")
			flightcache.Set(flightid+"_seen", []byte("seen"))
		}
	}
}

//@todo function based on distance from point and altitude
func is_visible_from(message sbsmessage.SBSMessage, pos_lat float64, pos_lon float64) (visibility bool, err error) {
	lat, lon, err := message.GetLatLong()
	if lat < 0 && lon > 0 {
		return true, nil
	} else {
		return false, nil
	}
}
