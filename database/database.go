package database

import (
	"database/sql"

	"github.com/Mowinski/LastWatchedBackend/logger"
)

var dbConn *sql.DB

// ConnectWithDatabase connect to selected mysql and associate it with global variable DBConn
func ConnectWithDatabase(dns string) (err error) {
	dbConn, err = sql.Open("mysql", dns)
	if err != nil {
		return err
	}
	return nil
}

// GetDBConn return database object
func GetDBConn() *sql.DB {
	err := dbConn.Ping()
	if err != nil {
		logger.Logger.Fatal("Can not ping database server, error: ", err)
	}

	return dbConn
}

// SetDBConn set database object
func SetDBConn(db *sql.DB) {
	dbConn = db
}
