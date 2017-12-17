package database

import (
	"database/sql"

	"github.com/Mowinski/LastWatchedBackend/logger"
)

var dbConn *sql.DB
var dnsConn string

// ConnectWithDatabase connect to selected mysql and associate it with global variable DBConn
func ConnectWithDatabase(dns string) (err error) {
	dbConn, err = sql.Open("mysql", dns)
	if err != nil {
		return err
	}
	dnsConn = dns
	return nil
}

// GetDBConn return database object
func GetDBConn() *sql.DB {
	if dbConn == nil {
		ConnectWithDatabase(dnsConn)
	}
	err := dbConn.Ping()
	if err != nil {
		logger.Logger.Fatal("Can not ping database server")
	}

	return dbConn
}
