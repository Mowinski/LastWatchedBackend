package movies

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Mowinski/LastWatchedBackend/logger"

	"github.com/Mowinski/LastWatchedBackend/database"
	"github.com/Mowinski/LastWatchedBackend/models"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var movieListRows *sqlmock.Rows

func setup(t *testing.T) sqlmock.Sqlmock {
	movieListRows = sqlmock.NewRows([]string{"id", "name", "url"}).
		AddRow(1, "Test Movie 1", "http://www.example.com/movie1").
		AddRow(2, "Test Movie 2", "http://www.example.com/movie2")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	database.SetDBConn(db)

	return mock
}
func TestRetriveMovieItems(t *testing.T) {
	mock := setup(t)

	mock.ExpectQuery("SELECT id, name, url FROM tv_series WHERE name LIKE (.+) LIMIT (.+) OFFSET (.+);").
		WithArgs("Test", 10, 0).
		WillReturnRows(movieListRows)

	movies, err := retriveMovieItems("Test", 10, 0)

	if err != nil {
		t.Errorf("Can no retrive movie items, got error: %s", err)
	}

	if len(movies) != 2 {
		t.Errorf("Wrong number of movies, expected 2, got %d", len(movies))
	}
}

func TestRetriveMovieItemsError(t *testing.T) {
	mock := setup(t)

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

func TestMovieListHandler(t *testing.T) {
	mock := setup(t)

	mock.ExpectQuery("SELECT id, name, url FROM tv_series WHERE name LIKE (.+) LIMIT (.+) OFFSET (.+);").
		WithArgs("%%", 50, 0).
		WillReturnRows(movieListRows)

	req, _ := http.NewRequest("GET", "/movies", nil)
	res := httptest.NewRecorder()

	MovieListHandler(res, req)

	if res.Code != 200 {
		t.Errorf("Wrong status code, expected 200, got %d", res.Code)
	}

	var movieList models.MovieItems
	json.Unmarshal(res.Body.Bytes(), &movieList)

	if movieList[0].ID != 1 {
		t.Errorf("Wrong response, ID expect 1, got %d", movieList[0].ID)
	}

	if movieList[1].ID != 2 {
		t.Errorf("Wrong response, ID expect 2, got %d", movieList[1].ID)
	}

	if movieList[0].Name != "Test Movie 1" {
		t.Errorf("Wrong response, Name expect 'Test Movie 1', got %s", movieList[0].Name)
	}

	if movieList[1].Name != "Test Movie 2" {
		t.Errorf("Wrong response, Name expect 'Test Movie 2', got %s", movieList[1].Name)
	}

	if movieList[0].URL != "http://www.example.com/movie1" {
		t.Errorf(
			"Wrong response, URL expect 'http://www.example.com/movie1', got %s",
			movieList[0].URL,
		)
	}

	if movieList[1].URL != "http://www.example.com/movie2" {
		t.Errorf(
			"Wrong response, URL expect 'http://www.example.com/movie2', got %s",
			movieList[1].URL,
		)
	}
}

func TestMovieListHandlerError(t *testing.T) {
	mock := setup(t)
	logger.SetLogger("test_log_file.txt")
	defer os.Remove("test_log_file.txt")

	mock.ExpectQuery("SELECT id, name, url FROM tv_series WHERE name LIKE (.+) LIMIT (.+) OFFSET (.+);").
		WithArgs("%%", 50, 0).
		WillReturnError(fmt.Errorf("Test error"))

	req, _ := http.NewRequest("GET", "/movies", nil)
	res := httptest.NewRecorder()

	MovieListHandler(res, req)

	if res.Code != 400 {
		t.Errorf("Wrong status code, expected 400, got %d", res.Code)
	}

	var errorMsg map[string]string
	json.Unmarshal(res.Body.Bytes(), &errorMsg)

	if errorMsg["error"] != "Test error" {
		t.Errorf("Wrong error message, expected 'Test error', got: %s", errorMsg["error"])
	}
}
