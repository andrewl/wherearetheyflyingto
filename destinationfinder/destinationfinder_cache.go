// Package destinationfinder provides functionality for
// determining where a flight with a given callsign is
// destined.
package destinationfinder

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

/**
 * Retrieves the lat long of the destination (as a simple string, we're not interested in doing
 * any real processing with this, just using it as an index.
 **/
type DestinationFinderCache struct {
	// pointer to sqlite3 db
	db *sql.DB
}

func (dfc DestinationFinderCache) Open(db *sql.DB) error {
	dfc.db = db

	create_table_sql := `
	    create table dest_cache (
				callsign text,
			  destination_lat_long text,
			`

	_, err := dfc.db.Exec(create_table_sql)
	if err != nil {
		//display a message?
		//logger.Log("msg", "create cache table failed. This probably isn't a problem if the table already exists.", "err", err)
	}
	return nil
}

func (dfc DestinationFinderCache) Cache_get(callsign string) string {
	if dfc.db == nil {
		return ""
	}

	var latlong string
	rows, err := dfc.db.Query("select destination_lat_long from dest_cache where callsign = '" + callsign + "'")
	if err != nil {
		_ = rows.Scan(&latlong)
	}
	return latlong
}

func (dfc DestinationFinderCache) Cache_set(callsign string, latlong string) {
	if dfc.db == nil {
		return
	}
	dfc.db.Exec("insert into dest_cache(callsign,destination_lat_long) values(?,?)", callsign, latlong)
}
