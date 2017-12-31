package movies

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Mowinski/LastWatchedBackend/database"
	"github.com/Mowinski/LastWatchedBackend/models"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type jsonBody string
type testStruct struct {
	TestID     int    `json:"testID"`
	TestString string `json:"testString"`
}

func (jb jsonBody) Read(p []byte) (n int, err error) {
	copy(p, jb)
	return len(jb), nil
}

func (jb jsonBody) Close() error {
	return nil
}

type movieTestInternalsData struct {
	movieListRows          *sqlmock.Rows
	movieDetailRow         *sqlmock.Rows
	movieDetailLastWatched *sqlmock.Rows
	validJSON              jsonBody
	invalidJSON            jsonBody
}

func setupInternals(t *testing.T) (*sql.DB, sqlmock.Sqlmock, movieTestInternalsData) {
	var testData movieTestInternalsData
	date, _ := time.Parse(time.RFC822Z, "2017-01-02 18:42:20")

	testData.movieListRows = sqlmock.NewRows([]string{"id", "name", "url"}).
		AddRow(1, "Test Movie 1", "http://www.example.com/movie1").
		AddRow(2, "Test Movie 2", "http://www.example.com/movie2")

	testData.movieDetailRow = sqlmock.NewRows([]string{"id", "name", "url", "seriesCount"}).
		AddRow(1, "Test Movie 1", "http://www.example.com/movie1", 5)
	testData.movieDetailLastWatched = sqlmock.NewRows([]string{"id", "id", "number", "date"}).
		AddRow(1, 1, 4, date)
	testData.validJSON = "{\"testID\":1,\"testString\":\"Test string\"}"
	testData.invalidJSON = "{\"testID\":1,testString: \"Test string with no quotation marks\"}"

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

	movies, err := retrieveMovieItems("Test", 10, 0)

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

	movies, err := retrieveMovieItems("Test", 10, 0)

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
	var movieHandler MovieHandlers

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

	movieDetail, err := movieHandler.CreateMovie(payload)

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

func TestCreateMovieFailTransaction(t *testing.T) {
	_, mock, _ := setupInternals(t)
	var movieHandler MovieHandlers

	mock.ExpectBegin().WillReturnError(fmt.Errorf("Transaction start error"))

	payload := models.MovieCreationPayload{
		MovieName:        "Test movie",
		URL:              "http://www.example.com",
		SeriesNumber:     2,
		EpisodesInSeries: 1,
	}

	movieDetail, err := movieHandler.CreateMovie(payload)

	if err.Error() != "Transaction start error" {
		t.Errorf("Wrong error, expected 'Transaction start error', got %s", err)
	}

	if movieDetail.ID != 0 {
		t.Error("Movie was created when error occure")
	}
}

func TestCreateMovieFailCreateSeries(t *testing.T) {
	_, mock, _ := setupInternals(t)
	var movieHandler MovieHandlers

	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO tv_series (.+)")
	mock.ExpectExec("(.)+").
		WillReturnError(fmt.Errorf("Test error durring create tv_series"))

	payload := models.MovieCreationPayload{
		MovieName:        "Test movie",
		URL:              "http://www.example.com",
		SeriesNumber:     2,
		EpisodesInSeries: 1,
	}

	movieDetail, err := movieHandler.CreateMovie(payload)

	if err.Error() != "Test error durring create tv_series" {
		t.Errorf("Wrong error, expected 'Test error durring create tv_series', got %s", err)
	}

	if movieDetail.ID != 0 {
		t.Error("Movie was created when error occure")
	}
}

func TestCreateMovieFailCreateSeason(t *testing.T) {
	_, mock, _ := setupInternals(t)
	var movieHandler MovieHandlers

	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO tv_series (.+)")
	mock.ExpectExec("(.)+").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectPrepare("INSERT INTO season (.+)")
	mock.ExpectExec("(.)+").
		WillReturnError(fmt.Errorf("Test error durring create season"))
	mock.ExpectRollback()

	payload := models.MovieCreationPayload{
		MovieName:        "Test movie",
		URL:              "http://www.example.com",
		SeriesNumber:     2,
		EpisodesInSeries: 1,
	}

	movieDetail, err := movieHandler.CreateMovie(payload)

	if err.Error() != "Test error durring create season" {
		t.Errorf("Wrong error, expected 'Test error durring create season', got %s", err)
	}

	if movieDetail.ID != 0 {
		t.Error("Movie was created when error occure")
	}
}

func TestCreateMovieFailCreateEpisode(t *testing.T) {
	_, mock, _ := setupInternals(t)
	var movieHandler MovieHandlers

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
		WillReturnError(fmt.Errorf("Test error during create episode"))

	mock.ExpectRollback()

	payload := models.MovieCreationPayload{
		MovieName:        "Test movie",
		URL:              "http://www.example.com",
		SeriesNumber:     2,
		EpisodesInSeries: 1,
	}

	movieDetail, err := movieHandler.CreateMovie(payload)

	if err.Error() != "Test error during create episode" {
		t.Errorf("Wrong error, expected 'Test error during create episode', got %s", err)
	}

	if movieDetail.ID != 0 {
		t.Error("Movie was created when error occure")
	}
}

func TestUpdateMovie(t *testing.T) {
	_, mock, testData := setupInternals(t)
	var movieHandler MovieHandlers

	mock.ExpectBegin()

	mock.ExpectPrepare("UPDATE tv_series SET (.+)")
	mock.ExpectExec("(.)+").
		WithArgs("Test movie", "http://www.example.com", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	mock.ExpectQuery("SELECT tv_series.id, tv_series.name, url(.+)").
		WithArgs(1).
		WillReturnRows(testData.movieDetailRow)

	mock.ExpectQuery("SELECT episode.id, season.id, episode.number, episode.date (.+)").
		WithArgs().
		WillReturnRows(testData.movieDetailLastWatched)

	payload := models.MovieUpdatePayload{
		MovieName:        "Test movie",
		URL:              "http://www.example.com",
		SeriesNumber:     2,
		EpisodesInSeries: 1,
	}

	movieDetail, err := movieHandler.UpdateMovie(1, payload)

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

func TestUpdateMovieFailBeginTransaction(t *testing.T) {
	_, mock, _ := setupInternals(t)
	var movieHandler MovieHandlers

	mock.ExpectBegin().WillReturnError(fmt.Errorf("Test error during begin"))

	payload := models.MovieUpdatePayload{
		MovieName:        "Test movie",
		URL:              "http://www.example.com",
		SeriesNumber:     2,
		EpisodesInSeries: 1,
	}

	movieDetail, err := movieHandler.UpdateMovie(1, payload)

	if err.Error() != "Test error during begin" {
		t.Errorf("Expected error 'Test error during begin', got %s", err)
	}

	if movieDetail.ID != 0 {
		t.Errorf("Wrong ID, expected 0, got %d", movieDetail.ID)
	}
}

func TestUpdateMovieFailExecuteUpdate(t *testing.T) {
	_, mock, _ := setupInternals(t)
	var movieHandler MovieHandlers

	mock.ExpectBegin()

	mock.ExpectPrepare("UPDATE tv_series SET (.+)")
	mock.ExpectExec("(.)+").
		WithArgs("Test movie", "http://www.example.com", 1).
		WillReturnError(fmt.Errorf("Test error during update"))

	payload := models.MovieUpdatePayload{
		MovieName:        "Test movie",
		URL:              "http://www.example.com",
		SeriesNumber:     2,
		EpisodesInSeries: 1,
	}

	movieDetail, err := movieHandler.UpdateMovie(1, payload)

	if err.Error() != "Test error during update" {
		t.Errorf("Expected error 'Test error during update', got %s", err)
	}

	if movieDetail.ID != 0 {
		t.Errorf("Wrong ID, expected 0, got %d", movieDetail.ID)
	}
}

func TestGetJSONParameters(t *testing.T) {
	_, _, testData := setupInternals(t)
	var out testStruct
	var movieHandler MovieHandlers

	err := movieHandler.GetJSONParameters(testData.validJSON, &out)

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if out.TestID != 1 {
		t.Errorf("Wrong testID, expected 1, got %d", out.TestID)
	}

	if out.TestString != "Test string" {
		t.Errorf("Wrong testString, expected 'Test string', got %s", out.TestString)
	}
}

func TestGetJSONParametersInvalidJSON(t *testing.T) {
	_, _, testData := setupInternals(t)
	var out testStruct
	var movieHandler MovieHandlers

	err := movieHandler.GetJSONParameters(testData.invalidJSON, &out)

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestRetrieveMovieDetail(t *testing.T) {
	_, mock, testData := setupInternals(t)
	var movieHandler MovieHandlers

	mock.ExpectQuery("SELECT tv_series(.+)").
		WithArgs(1).
		WillReturnRows(testData.movieDetailRow)

	mock.ExpectQuery("SELECT episode(.+)").
		WithArgs(1).
		WillReturnRows(testData.movieDetailLastWatched)

	movie, err := movieHandler.RetrieveMovieDetail(1)

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if movie.ID != 1 {
		t.Errorf("Wrong movie ID, expected 1, got %d", movie.ID)
	}
}

func TestRetrieveMovieDetailFailTVSeriesQuery(t *testing.T) {
	_, mock, _ := setupInternals(t)
	var movieHandler MovieHandlers

	mock.ExpectQuery("SELECT tv_series(.+)").
		WithArgs(1).
		WillReturnError(fmt.Errorf("Test error during tv_series"))

	movie, err := movieHandler.RetrieveMovieDetail(1)

	if err.Error() != "Test error during tv_series" {
		t.Errorf("Expected error 'Test error during tv_series', got: %s", err)
	}

	if movie.ID != 0 {
		t.Errorf("Wrong movie ID, expected 0, got %d", movie.ID)
	}
}

func TestRetrieveMovieDetailFailEpisodeQuery(t *testing.T) {
	_, mock, testData := setupInternals(t)
	var movieHandler MovieHandlers

	mock.ExpectQuery("SELECT tv_series(.+)").
		WithArgs(1).
		WillReturnRows(testData.movieDetailRow)

	mock.ExpectQuery("SELECT episode(.+)").
		WithArgs(1).
		WillReturnError(fmt.Errorf("Test error during episode"))

	movie, err := movieHandler.RetrieveMovieDetail(1)

	if err != nil {
		t.Errorf("Unexpected error, got: %s", err)
	}

	if movie.ID != 1 {
		t.Errorf("Wrong movie ID, expected 1, got %d", movie.ID)
	}
}

func TestDeleteMovie(t *testing.T) {
	_, mock, _ := setupInternals(t)

	var movieHandler MovieHandlers

	mock.ExpectPrepare("DELETE FROM tv_series(.+)")
	mock.ExpectExec("(.+)").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := movieHandler.DeleteMovie(1)

	if err != nil {
		t.Errorf("Unexpected error, got %s", err)
	}
}

func TestDeleteMovieFailedPrepare(t *testing.T) {
	_, mock, _ := setupInternals(t)

	var movieHandler MovieHandlers

	mock.ExpectPrepare("DELETE FROM tv_series(.+)").
		WillReturnError(fmt.Errorf("Test error during prepare"))

	err := movieHandler.DeleteMovie(1)

	if err.Error() != "Test error during prepare" {
		t.Errorf("Expected error 'Test error during prepare', got %s", err)
	}
}

func TestDeleteMovieFailedExecute(t *testing.T) {
	_, mock, _ := setupInternals(t)

	var movieHandler MovieHandlers

	mock.ExpectPrepare("DELETE FROM tv_series(.+)")
	mock.ExpectExec("(.+)").
		WithArgs(1).
		WillReturnError(fmt.Errorf("Test error in execute"))

	err := movieHandler.DeleteMovie(1)

	if err.Error() != "Test error in execute" {
		t.Errorf("Expected error 'Test error in execute', got %s", err)
	}
}
