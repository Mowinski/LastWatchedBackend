package database

import (
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestSetUpDatabaseConnection(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	SetDBConn(db)

	if db != dbConn {
		t.Error("Seting DB connection does not set connection")
	}
}

func TestGetDBConnection(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	SetDBConn(db)

	if db != GetDBConn() {
		t.Error("Set DB object does not match with getting")
	}
}

func TestConnectWithDB(t *testing.T) {
	err := ConnectWithDatabase("notexist")
	if err == nil {
		t.Error("Expected exception, got nil")
	}
}
