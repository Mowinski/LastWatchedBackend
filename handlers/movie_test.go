package movies

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Mowinski/LastWatchedBackend/logger"
	"github.com/gorilla/mux"

	"github.com/Mowinski/LastWatchedBackend/database"
	"github.com/Mowinski/LastWatchedBackend/models"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type movieBodyPayload string
type MovieUtilsMocked struct{}

func (m movieBodyPayload) Read(p []byte) (n int, err error) {
	copy(p, m)
	return len(m), nil
}
func (m movieBodyPayload) Close() error {
	return nil
}

func (mh MovieUtilsMocked) createMovie(payload models.MovieCreationPayload) (movie models.MovieDetail, err error) {
	movie.ID = 1
	movie.Name = "Test movie"
	movie.URL = "http://www.example.com/test-movie"
	movie.DateOfLastWatchedEpisode = time.Now()
	movie.LastWatchedEpisode.ID = 2
	movie.LastWatchedEpisode.Series = 3
	movie.LastWatchedEpisode.EpisodeNumber = 3
	return movie, nil
}

var movieListRows *sqlmock.Rows
var movieDetailRow *sqlmock.Rows
var movieDetailLastWatched *sqlmock.Rows
var movieCreatePayload movieBodyPayload
var movieHandlers MovieHandlers

func setup(t *testing.T) sqlmock.Sqlmock {
	movieListRows = sqlmock.NewRows([]string{"id", "name", "url"}).
		AddRow(1, "Test Movie 1", "http://www.example.com/movie1").
		AddRow(2, "Test Movie 2", "http://www.example.com/movie2")

	movieDetailRow = sqlmock.NewRows([]string{"id", "name", "url", "seriesCount"}).
		AddRow(1, "Test Movie 1", "http://www.example.com/movie1", 5)
	movieDetailLastWatched = sqlmock.NewRows([]string{"id", "id", "number", "date"}).
		AddRow(1, 1, 4, "2017-01-02 18:42:20")
	movieCreatePayload = "{\"movieName\":\"Marvel Runaways\",\"url\":\"www.google.com/url\",\"seriesNumber\":1,\"episodesInSeries\":10}"

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	var mh MovieUtilsMocked
	movieHandlers.utils = mh

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

	movieHandlers.MovieListHandler(res, req)

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

	movieHandlers.MovieListHandler(res, req)

	if res.Code != 400 {
		t.Errorf("Wrong status code, expected 400, got %d", res.Code)
	}

	var errorMsg map[string]string
	json.Unmarshal(res.Body.Bytes(), &errorMsg)

	if errorMsg["error"] != "Test error" {
		t.Errorf("Wrong error message, expected 'Test error', got: %s", errorMsg["error"])
	}
}

func TestMovieDetailsHanlder(t *testing.T) {
	mock := setup(t)

	mock.ExpectQuery("SELECT tv_series.id, tv_series.name, url(.+)").
		WithArgs(1).
		WillReturnRows(movieDetailRow)

	mock.ExpectQuery("SELECT episode.id, season.id, episode.number, episode.date (.+)").
		WithArgs().
		WillReturnRows(movieDetailLastWatched)

	req, _ := http.NewRequest("GET", "/movie/1", nil)
	res := httptest.NewRecorder()

	m := mux.NewRouter()
	m.HandleFunc("/movie/{id}", movieHandlers.MovieDetailsHanlder).Methods("GET")
	m.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Wrong status code, expected 200, got %d", res.Code)
	}

	var movieDetail models.MovieDetail
	json.Unmarshal(res.Body.Bytes(), &movieDetail)

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

func TestMovieDetailsHanlderError(t *testing.T) {
	mock := setup(t)

	mock.ExpectQuery("(.+)").
		WithArgs(1).
		WillReturnError(fmt.Errorf("Test error"))

	req, _ := http.NewRequest("GET", "/movie/1", nil)
	res := httptest.NewRecorder()

	m := mux.NewRouter()
	m.HandleFunc("/movie/{id}", movieHandlers.MovieDetailsHanlder).Methods("GET")
	m.ServeHTTP(res, req)

	if res.Code != 400 {
		t.Errorf("Wrong status code, expected 400, got %d", res.Code)
	}

	var errorMsg map[string]string
	json.Unmarshal(res.Body.Bytes(), &errorMsg)

	if errorMsg["error"] != "Test error" {
		t.Errorf("Wrong error message, expected 'Test error', got: %s", errorMsg["error"])
	}
}

func TestMovieCreateHandler(t *testing.T) {
	setup(t)
	req, _ := http.NewRequest("POST", "/movie", nil)
	res := httptest.NewRecorder()
	req.Body = movieCreatePayload

	movieHandlers.MovieCreateHandler(res, req)

	if res.Code != 200 {
		t.Errorf("Wrong status code, expected 200, got %d", res.Code)
	}

	var movieDetail models.MovieDetail
	json.Unmarshal(res.Body.Bytes(), &movieDetail)

	if movieDetail.ID != 1 {
		t.Errorf("Wrong movie id, expected 1, got %d", movieDetail.ID)
	}
	if movieDetail.Name != "Test movie" {
		t.Errorf("Wrong movie name, expected 'Test movie', got %s", movieDetail.Name)
	}
	if movieDetail.URL != "http://www.example.com/test-movie" {
		t.Errorf("Wrong movie URL, expected 'http://www.example.com/test-movie', got %s", movieDetail.URL)
	}
	if movieDetail.LastWatchedEpisode.ID != 2 {
		t.Errorf("Wrong last watched episode ID, expected 2, got %d", movieDetail.LastWatchedEpisode.ID)
	}
	if movieDetail.LastWatchedEpisode.Series != 3 {
		t.Errorf("Wrong last watched episode series, expected 3, got %d", movieDetail.LastWatchedEpisode.Series)
	}
	if movieDetail.LastWatchedEpisode.EpisodeNumber != 3 {
		t.Errorf("Wrong last watched episode, expected 3, got %d", movieDetail.LastWatchedEpisode.EpisodeNumber)
	}
}
