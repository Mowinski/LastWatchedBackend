package movies_test

import (
	"encoding/json"
	"fmt"
	"github.com/Mowinski/LastWatchedBackend/database"
	"github.com/Mowinski/LastWatchedBackend/logger"
	"github.com/Mowinski/LastWatchedBackend/models"
	"github.com/Mowinski/LastWatchedBackend/utils"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Mowinski/LastWatchedBackend/handlers"
	"github.com/gorilla/mux"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type movieBodyPayload string

type MovieUtilsSuccessMocked struct{}
type MovieUtilsCreateFailedMocked struct{}
type MovieUtilsJSONParseFailedMocked struct{}

type movieTestHandlerData struct {
	movieListRows          *sqlmock.Rows
	movieDetailRow         *sqlmock.Rows
	movieDetailLastWatched *sqlmock.Rows

	movieCreatePayload           movieBodyPayload
	movieSuccessHandlers         movies.MovieHandlers
	movieCreateFailedHandlers    movies.MovieHandlers
	movieJSONParseFailedHandlers movies.MovieHandlers
}

func (m movieBodyPayload) Read(p []byte) (n int, err error) {
	copy(p, m)
	return len(m), nil
}
func (m movieBodyPayload) Close() error {
	return nil
}

func (mh MovieUtilsSuccessMocked) CreateMovie(payload models.MovieCreationPayload) (movie models.MovieDetail, err error) {
	movie.ID = 1
	movie.Name = "Test movie"
	movie.URL = "http://www.example.com/test-movie"
	movie.DateOfLastWatchedEpisode, _ = time.Parse(time.RFC822Z, "29 Jan 91 03:04 -0700")
	movie.LastWatchedEpisode.ID = 2
	movie.LastWatchedEpisode.Series = 3
	movie.LastWatchedEpisode.EpisodeNumber = 3
	return movie, nil
}

func (mh MovieUtilsCreateFailedMocked) CreateMovie(payload models.MovieCreationPayload) (movie models.MovieDetail, err error) {
	return movie, fmt.Errorf("Test error durring create movie")
}

func (mh MovieUtilsJSONParseFailedMocked) CreateMovie(payload models.MovieCreationPayload) (movie models.MovieDetail, err error) {
	return movie, fmt.Errorf("Test error durring create movie")
}

func (mh MovieUtilsSuccessMocked) GetJSONParameters(body io.ReadCloser, out interface{}) error {
	return utils.GetJSONParameters(body, out)
}

func (mh MovieUtilsCreateFailedMocked) GetJSONParameters(body io.ReadCloser, out interface{}) error {
	return utils.GetJSONParameters(body, out)
}

func (mh MovieUtilsJSONParseFailedMocked) GetJSONParameters(body io.ReadCloser, out interface{}) error {
	return fmt.Errorf("Test error during parsing JSON parameters")
}

func setup(t *testing.T) (sqlmock.Sqlmock, movieTestHandlerData) {
	var testData movieTestHandlerData

	testData.movieListRows = sqlmock.NewRows([]string{"id", "name", "url"}).
		AddRow(1, "Test Movie 1", "http://www.example.com/movie1").
		AddRow(2, "Test Movie 2", "http://www.example.com/movie2")

	testData.movieDetailRow = sqlmock.NewRows([]string{"id", "name", "url", "seriesCount"}).
		AddRow(1, "Test Movie 1", "http://www.example.com/movie1", 5)
	testData.movieDetailLastWatched = sqlmock.NewRows([]string{"id", "id", "number", "date"}).
		AddRow(1, 1, 4, "2017-01-02 18:42:20")
	testData.movieCreatePayload = "{\"movieName\":\"Marvel Runaways\",\"url\":\"www.google.com/url\",\"seriesNumber\":1,\"episodesInSeries\":10}"

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	var successUtils MovieUtilsSuccessMocked
	var failedUtils MovieUtilsCreateFailedMocked
	var jsonFailedUtils MovieUtilsJSONParseFailedMocked
	testData.movieSuccessHandlers = movies.MovieHandlers{Utils: successUtils}
	testData.movieCreateFailedHandlers = movies.MovieHandlers{Utils: failedUtils}
	testData.movieJSONParseFailedHandlers = movies.MovieHandlers{Utils: jsonFailedUtils}
	database.SetDBConn(db)

	return mock, testData
}

func TestMovieListHandler(t *testing.T) {
	mock, testData := setup(t)

	mock.ExpectQuery("SELECT id, name, url FROM tv_series WHERE name LIKE (.+) LIMIT (.+) OFFSET (.+);").
		WithArgs("%%", 50, 0).
		WillReturnRows(testData.movieListRows)

	req, _ := http.NewRequest("GET", "/movies", nil)
	res := httptest.NewRecorder()

	testData.movieSuccessHandlers.MovieListHandler(res, req)

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
	mock, testData := setup(t)
	logger.SetLogger("test_log_file.txt")
	defer os.Remove("test_log_file.txt")

	mock.ExpectQuery("SELECT id, name, url FROM tv_series WHERE name LIKE (.+) LIMIT (.+) OFFSET (.+);").
		WithArgs("%%", 50, 0).
		WillReturnError(fmt.Errorf("Test error"))

	req, _ := http.NewRequest("GET", "/movies", nil)
	res := httptest.NewRecorder()

	testData.movieSuccessHandlers.MovieListHandler(res, req)

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
	mock, testData := setup(t)
	logger.SetLogger("test_log_file.txt")
	defer os.Remove("test_log_file.txt")

	mock.ExpectQuery("SELECT tv_series.id, tv_series.name, url(.+)").
		WithArgs(1).
		WillReturnRows(testData.movieDetailRow)

	mock.ExpectQuery("SELECT episode.id, season.id, episode.number, episode.date (.+)").
		WithArgs().
		WillReturnRows(testData.movieDetailLastWatched)

	req, _ := http.NewRequest("GET", "/movie/1", nil)
	res := httptest.NewRecorder()

	m := mux.NewRouter()
	m.HandleFunc("/movie/{id}", testData.movieSuccessHandlers.MovieDetailsHanlder).Methods("GET")
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
	mock, testData := setup(t)
	logger.SetLogger("test_log_file.txt")
	defer os.Remove("test_log_file.txt")

	mock.ExpectQuery("(.+)").
		WithArgs(1).
		WillReturnError(fmt.Errorf("Test error"))

	req, _ := http.NewRequest("GET", "/movie/1", nil)
	res := httptest.NewRecorder()

	m := mux.NewRouter()
	m.HandleFunc("/movie/{id}", testData.movieSuccessHandlers.MovieDetailsHanlder).Methods("GET")
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
	_, testData := setup(t)
	req, _ := http.NewRequest("POST", "/movie", nil)
	res := httptest.NewRecorder()
	req.Body = testData.movieCreatePayload

	testData.movieSuccessHandlers.MovieCreateHandler(res, req)

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
	year, month, day := movieDetail.DateOfLastWatchedEpisode.Date()
	if year != 1991 || month != 1 || day != 29 {
		t.Errorf("Wrong date of last watched, expected 1991-01-29, got %v", movieDetail.DateOfLastWatchedEpisode)
	}

	if movieDetail.DateOfLastWatchedEpisode.Unix() != 665143440 {
		t.Errorf("Wrong date and time of last watched, expected timestamp 665143440, got %d", movieDetail.DateOfLastWatchedEpisode.Unix())
	}
}

func TestMovieCreateErrorHandler(t *testing.T) {
	_, testData := setup(t)
	req, _ := http.NewRequest("POST", "/movie", nil)
	res := httptest.NewRecorder()
	req.Body = testData.movieCreatePayload

	testData.movieCreateFailedHandlers.MovieCreateHandler(res, req)

	if res.Code != 400 {
		t.Errorf("Wrong status code, expected 400, got %d", res.Code)
	}

	var errorMsg map[string]string
	json.Unmarshal(res.Body.Bytes(), &errorMsg)

	if errorMsg["error"] != "Test error durring create movie" {
		t.Errorf("Wrong error message, expected 'Test error durring create movie', got: %s", errorMsg["error"])
	}
}

func TestMovieCreateParseJSONErrorHandler(t *testing.T) {
	_, testData := setup(t)
	req, _ := http.NewRequest("POST", "/movie", nil)
	res := httptest.NewRecorder()
	req.Body = testData.movieCreatePayload

	testData.movieJSONParseFailedHandlers.MovieCreateHandler(res, req)

	if res.Code != 400 {
		t.Errorf("Wrong status code, expected 400, got %d", res.Code)
	}

	var errorMsg map[string]string
	json.Unmarshal(res.Body.Bytes(), &errorMsg)

	if errorMsg["error"] != "Test error during parsing JSON parameters" {
		t.Errorf("Wrong error message, expected 'Test error during parsing JSON parameters', got: %s", errorMsg["error"])
	}
}
