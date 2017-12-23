package movies

import (
	"fmt"
	"testing"

	"github.com/Mowinski/LastWatchedBackend/database"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type movieTestInternalsData struct {
	movieListRows *sqlmock.Rows
}

func setupInternals(t *testing.T) (sqlmock.Sqlmock, movieTestInternalsData) {
	var testData movieTestInternalsData

	testData.movieListRows = sqlmock.NewRows([]string{"id", "name", "url"}).
		AddRow(1, "Test Movie 1", "http://www.example.com/movie1").
		AddRow(2, "Test Movie 2", "http://www.example.com/movie2")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	database.SetDBConn(db)

	return mock, testData
}
func TestRetriveMovieItems(t *testing.T) {
	mock, testData := setupInternals(t)

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
	mock, _ := setupInternals(t)

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
