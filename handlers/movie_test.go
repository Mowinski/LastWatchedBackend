package movies_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Mowinski/LastWatchedBackend/database"
	"github.com/Mowinski/LastWatchedBackend/logger"
	"github.com/Mowinski/LastWatchedBackend/models"

	"github.com/Mowinski/LastWatchedBackend/handlers"
	"github.com/gorilla/mux"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func setup(t *testing.T) (sqlmock.Sqlmock, movieTestHandlerData) {
	var testData movieTestHandlerData
	date, _ := time.Parse(time.RFC822Z, "2017-01-02 18:42:20")

	testData.movieListRows = sqlmock.NewRows([]string{"id", "name", "url"}).
		AddRow(1, "Test Movie 1", "http://www.example.com/movie1").
		AddRow(2, "Test Movie 2", "http://www.example.com/movie2")

	testData.movieDetailRow = sqlmock.NewRows([]string{"id", "name", "url", "seriesCount"}).
		AddRow(1, "Test Movie 1", "http://www.example.com/movie1", 5)
	testData.movieDetailLastWatched = sqlmock.NewRows([]string{"id", "id", "number", "date"}).
		AddRow(1, 1, 4, date)
	testData.movieCreatePayload = "{\"movieName\":\"Marvel Runaways\",\"url\":\"www.google.com/url\",\"seriesNumber\":1,\"episodesInSeries\":10}"
	testData.movieUpdatePayload = "{\"movieName\":\"Marvel Runaways New\",\"url\":\"www.google.com/new-url\",\"seriesNumber\": 1,\"episodesInSeries\": 10}"
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	var successUtils MovieUtilsSuccessMocked
	var failedUtils MovieUtilsCreateFailedMocked
	var jsonFailedUtils MovieUtilsJSONParseFailedMocked
	var jsonUpdateFailedUtils MovieUtilsUpdateMovieFailedMocked
	var movieDeleteFailedUtils MovieUtilsDeleteMovieFailedMocked
	var movieRetrieveFailedUtils MovieRetrieveDetailFailedMocked

	testData.movieSuccessHandlers = movies.MovieHandlers{Utils: successUtils}
	testData.movieCreateFailedHandlers = movies.MovieHandlers{Utils: failedUtils}
	testData.movieJSONParseFailedHandlers = movies.MovieHandlers{Utils: jsonFailedUtils}
	testData.movieUpdateFailedHandlers = movies.MovieHandlers{Utils: jsonUpdateFailedUtils}
	testData.movieDeleteFailedHandlers = movies.MovieHandlers{Utils: movieDeleteFailedUtils}
	testData.movieRetrieveDetailFailedHandlers = movies.MovieHandlers{Utils: movieRetrieveFailedUtils}
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

func TestMovieDetailsHandler(t *testing.T) {
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
	m.HandleFunc("/movie/{id}", testData.movieSuccessHandlers.MovieDetailsHandler).Methods("GET")
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

func TestMovieDetailsHandlerError(t *testing.T) {
	_, testData := setup(t)
	logger.SetLogger("test_log_file.txt")
	defer os.Remove("test_log_file.txt")

	req, _ := http.NewRequest("GET", "/movie/1", nil)
	res := httptest.NewRecorder()

	m := mux.NewRouter()
	m.HandleFunc("/movie/{id}", testData.movieRetrieveDetailFailedHandlers.MovieDetailsHandler).Methods("GET")
	m.ServeHTTP(res, req)

	if res.Code != 400 {
		t.Errorf("Wrong status code, expected 400, got %d", res.Code)
	}

	var errorMsg map[string]string
	json.Unmarshal(res.Body.Bytes(), &errorMsg)

	if errorMsg["error"] != "Test error during retrieve" {
		t.Errorf("Wrong error message, expected 'Test error during retrieve', got: %s", errorMsg["error"])
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
	req, _ := http.NewRequest("POST", "/movie", testData.movieCreatePayload)
	res := httptest.NewRecorder()

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

func TestMovieUpdateHandler(t *testing.T) {
	_, testData := setup(t)
	req, _ := http.NewRequest("PUT", "/movie/1", testData.movieUpdatePayload)
	res := httptest.NewRecorder()

	m := mux.NewRouter()
	m.HandleFunc("/movie/{id}", testData.movieSuccessHandlers.MovieUpdateHandler).Methods("PUT")
	m.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Wrong status code, expected 200, got %d", res.Code)
	}

	var movieDetail models.MovieDetail
	json.Unmarshal(res.Body.Bytes(), &movieDetail)

	if movieDetail.ID != 1 {
		t.Errorf("Wrong movie id, expected 1, got %d", movieDetail.ID)
	}

	if movieDetail.Name != "Marvel Runaways New" {
		t.Errorf("Wrong movie name, expected 'Marvel Runaways New', got %s", movieDetail.Name)
	}

	if movieDetail.URL != "www.google.com/new-url" {
		t.Errorf("Wrong movie URL, expected 'www.google.com/new-url', got %s", movieDetail.URL)
	}
}

func TestMovieUpdateFailParametersHandler(t *testing.T) {
	_, testData := setup(t)
	req, _ := http.NewRequest("PUT", "/movie/1", testData.movieUpdatePayload)
	res := httptest.NewRecorder()

	m := mux.NewRouter()
	m.HandleFunc("/movie/{id}", testData.movieJSONParseFailedHandlers.MovieUpdateHandler).Methods("PUT")
	m.ServeHTTP(res, req)

	if res.Code != 400 {
		t.Errorf("Wrong status code, expected 400, got %d", res.Code)
	}

	var errorMsg map[string]string
	json.Unmarshal(res.Body.Bytes(), &errorMsg)

	if errorMsg["error"] != "Test error during parsing JSON parameters" {
		t.Errorf("Wrong error message, expected 'Test error during parsing JSON parameters', got: %s", errorMsg["error"])
	}
}

func TestMovieUpdateFailUpdateHandler(t *testing.T) {
	_, testData := setup(t)
	req, _ := http.NewRequest("PUT", "/movie/1", testData.movieUpdatePayload)
	res := httptest.NewRecorder()

	m := mux.NewRouter()
	m.HandleFunc("/movie/{id}", testData.movieUpdateFailedHandlers.MovieUpdateHandler).Methods("PUT")
	m.ServeHTTP(res, req)

	if res.Code != 400 {
		t.Errorf("Wrong status code, expected 400, got %d", res.Code)
	}

	var errorMsg map[string]string
	json.Unmarshal(res.Body.Bytes(), &errorMsg)

	if errorMsg["error"] != "Test error during update movie" {
		t.Errorf("Wrong error message, expected 'Test error during update movie', got: %s", errorMsg["error"])
	}
}

func TestMovieUpdateFailRetrieveMovieHandler(t *testing.T) {
	_, testData := setup(t)
	req, _ := http.NewRequest("PUT", "/movie/1", testData.movieUpdatePayload)
	res := httptest.NewRecorder()

	m := mux.NewRouter()
	m.HandleFunc("/movie/{id}", testData.movieRetrieveDetailFailedHandlers.MovieUpdateHandler).Methods("PUT")
	m.ServeHTTP(res, req)

	if res.Code != 404 {
		t.Errorf("Wrong status code, expected 404, got %d", res.Code)
	}

	var errorMsg map[string]string
	json.Unmarshal(res.Body.Bytes(), &errorMsg)

	if errorMsg["error"] != "" {
		t.Errorf("Wrong error message, expected '', got: %s", errorMsg["error"])
	}
}

func TestMovieDeleteHandler(t *testing.T) {
	_, testData := setup(t)

	req, _ := http.NewRequest("DELETE", "/movie/1", nil)
	res := httptest.NewRecorder()

	m := mux.NewRouter()
	m.HandleFunc("/movie/{id}", testData.movieSuccessHandlers.MovieDeleteHandler).Methods("DELETE")
	m.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Wrong status code, expected 200, got %d", res.Code)
	}
}

func TestMovieDeleteFailedHandler(t *testing.T) {
	_, testData := setup(t)

	req, _ := http.NewRequest("DELETE", "/movie/1", nil)
	res := httptest.NewRecorder()

	m := mux.NewRouter()
	m.HandleFunc("/movie/{id}", testData.movieDeleteFailedHandlers.MovieDeleteHandler).Methods("DELETE")
	m.ServeHTTP(res, req)

	if res.Code != 400 {
		t.Errorf("Wrong status code, expected 400, got %d", res.Code)
	}

	var errorMsg map[string]string
	json.Unmarshal(res.Body.Bytes(), &errorMsg)

	if errorMsg["error"] != "Test error during delete movie" {
		t.Errorf("Wrong error message, expected 'Test error during delete movie', got: %s", errorMsg["error"])
	}
}

func TestMovieDeleteRetrieveFailedHandler(t *testing.T) {
	_, testData := setup(t)

	req, _ := http.NewRequest("DELETE", "/movie/1", nil)
	res := httptest.NewRecorder()

	m := mux.NewRouter()
	m.HandleFunc("/movie/{id}", testData.movieRetrieveDetailFailedHandlers.MovieDeleteHandler).Methods("DELETE")
	m.ServeHTTP(res, req)

	if res.Code != 404 {
		t.Errorf("Wrong status code, expected 404, got %d", res.Code)
	}

	var errorMsg map[string]string
	json.Unmarshal(res.Body.Bytes(), &errorMsg)

	if errorMsg["error"] != "" {
		t.Errorf("Wrong error message, expected '', got: %s", errorMsg["error"])
	}
}
