package movies

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/Mowinski/LastWatchedBackend/database"
	"github.com/Mowinski/LastWatchedBackend/models"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type movieTestInternalsData struct {
	movieListRows          *sqlmock.Rows
	movieDetailRow         *sqlmock.Rows
	movieDetailLastWatched *sqlmock.Rows
}

func setupInternals(t *testing.T) (*sql.DB, sqlmock.Sqlmock, movieTestInternalsData) {
	var testData movieTestInternalsData

	testData.movieListRows = sqlmock.NewRows([]string{"id", "name", "url"}).
		AddRow(1, "Test Movie 1", "http://www.example.com/movie1").
		AddRow(2, "Test Movie 2", "http://www.example.com/movie2")

	testData.movieDetailRow = sqlmock.NewRows([]string{"id", "name", "url", "seriesCount"}).
		AddRow(1, "Test Movie 1", "http://www.example.com/movie1", 5)
	testData.movieDetailLastWatched = sqlmock.NewRows([]string{"id", "id", "number", "date"}).
		AddRow(1, 1, 4, "2017-01-02 18:42:20")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	database.SetDBConn(db)

	return db, mock, testData
}
func TestRetriveMovieItems(t *testing.T) {
	_, mock, testData := setupInternals(t)

	mock.ExpectQuery("SELECT id, name, url FROM tv_series WHERE name LIKE (.+) LIMIT (.+) OFFSET (.+);").
		WithArgs("Test", 10, 0).
		WillReturnRows(testData.movieListRows)

	movies, err := retriveMovieItems("Test", 10, 0)

	if err != nil {
		t.Errorf("Can no retrive movie items, got error: %s", err)
	}

	if len(movies) != 2 {
		t.Errorf("Wrong number of movies, expected 2, got %d", len(movies))
	}
}

func TestRetriveMovieItemsError(t *testing.T) {
	_, mock, _ := setupInternals(t)

	mock.ExpectQuery("SELECT id, name, url FROM tv_series WHERE name LIKE (.+) LIMIT (.+) OFFSET (.+);").
		WithArgs("Test", 10, 0).
		WillReturnError(fmt.Errorf("Test Error"))

	movies, err := retriveMovieItems("Test", 10, 0)

	if err == nil {
		t.Errorf("Function does not return error")
	}

	if len(movies) != 0 {
		t.Errorf("Wrong number of movies, expected 0, got %d", len(movies))
	}
}

func TestExecuteStmtPrepareError(t *testing.T) {
	db, mock, _ := setupInternals(t)

	mock.ExpectBegin()
	mock.ExpectPrepare("(.)+").
		WillReturnError(fmt.Errorf("Test error in prepare"))

	tx, _ := db.Begin()

	id, err := executeStmt(tx, "SELECT * FROM test")

	if id != 0 {
		t.Errorf("Wrong id, expected 0, got %d", id)
	}

	if err.Error() != "Test error in prepare" {
		t.Errorf("Wrong error message, expected 'Test error in prepare', got %s", err)
	}
}

func TestExecuteStmtExecError(t *testing.T) {
	db, mock, _ := setupInternals(t)

	mock.ExpectBegin()
	mock.ExpectPrepare("(.+)")
	mock.ExpectExec("(.+)").
		WillReturnError(fmt.Errorf("Test error in execute"))

	tx, _ := db.Begin()

	id, err := executeStmt(tx, "SELECT * FROM test")

	if id != 0 {
		t.Errorf("Wrong id, expected 0, got %d", id)
	}

	if err.Error() != "Test error in execute" {
		t.Errorf("Wrong error message, expected 'Test error in execute', got %s", err)
	}
}

func TestExecuteStmt(t *testing.T) {
	db, mock, _ := setupInternals(t)

	mock.ExpectBegin()
	mock.ExpectPrepare("(.+)")
	mock.ExpectExec("(.+)").
		WillReturnResult(sqlmock.NewResult(1, 2))

	tx, _ := db.Begin()

	id, err := executeStmt(tx, "SELECT * FROM test")

	if id != 1 {
		t.Errorf("Wrong id, expected 1, got %d", id)
	}

	if err != nil {
		t.Errorf("Wrong error message, expected 'nil', got %s", err)
	}
}

func TestCreateMovie(t *testing.T) {
	_, mock, testData := setupInternals(t)

	mock.ExpectBegin()
	// Create movie
	mock.ExpectPrepare("INSERT INTO tv_series (.+)")
	mock.ExpectExec("(.)+").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Create season 1
	mock.ExpectPrepare("INSERT INTO season (.+)")
	mock.ExpectExec("(.)+").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Create episode
	mock.ExpectPrepare("INSERT INTO episode (.+)")
	mock.ExpectExec("(.)+").
		WithArgs(1, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create season 2
	mock.ExpectPrepare("INSERT INTO season (.+)")
	mock.ExpectExec("(.)+").
		WillReturnResult(sqlmock.NewResult(2, 1))
	// Create episode
	mock.ExpectPrepare("INSERT INTO episode (.+)")
	mock.ExpectExec("(.)+").
		WithArgs(2, 1).
		WillReturnResult(sqlmock.NewResult(2, 1))

	mock.ExpectCommit()

	mock.ExpectQuery("SELECT tv_series.id, tv_series.name, url(.+)").
		WithArgs(1).
		WillReturnRows(testData.movieDetailRow)

	mock.ExpectQuery("SELECT episode.id, season.id, episode.number, episode.date (.+)").
		WithArgs().
		WillReturnRows(testData.movieDetailLastWatched)

	payload := models.MovieCreationPayload{
		MovieName:        "Test movie",
		URL:              "http://www.example.com",
		SeriesNumber:     2,
		EpisodesInSeries: 1,
	}

	movieDetail, err := createMovie(payload)

	if err != nil {
		t.Errorf("Unexpected error, got %s", err)
	}

	if movieDetail.ID != 1 {
		t.Errorf("Wrong ID, expected 1, got %d", movieDetail.ID)
	}

	if movieDetail.Name != "Test Movie 1" {
		t.Errorf("Wrong movie name, expected 'Test Movie 1', got %s", movieDetail.Name)
	}

	if movieDetail.URL != "http://www.example.com/movie1" {
		t.Errorf("Wrong movie url, expected 'http://www.example.com/movie1', got %s", movieDetail.URL)
	}

	if movieDetail.SeriesCount != 5 {
		t.Errorf("Wrong movie series count, expected 5, got %d", movieDetail.SeriesCount)
	}
}
