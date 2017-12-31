package movies_test

import (
	"fmt"
	"io"
	"time"

	"github.com/Mowinski/LastWatchedBackend/handlers"
	"github.com/Mowinski/LastWatchedBackend/models"
	"github.com/Mowinski/LastWatchedBackend/utils"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type movieBodyPayload string

type movieTestHandlerData struct {
	movieListRows          *sqlmock.Rows
	movieDetailRow         *sqlmock.Rows
	movieDetailLastWatched *sqlmock.Rows

	movieCreatePayload                movieBodyPayload
	movieUpdatePayload                movieBodyPayload
	movieSuccessHandlers              movies.MovieHandlers
	movieCreateFailedHandlers         movies.MovieHandlers
	movieJSONParseFailedHandlers      movies.MovieHandlers
	movieUpdateFailedHandlers         movies.MovieHandlers
	movieDeleteFailedHandlers         movies.MovieHandlers
	movieRetrieveDetailFailedHandlers movies.MovieHandlers
}

func (m movieBodyPayload) Read(p []byte) (n int, err error) {
	copy(p, m)
	return len(m), nil
}
func (m movieBodyPayload) Close() error {
	return nil
}

// MovieUtilsSuccessMocked
type MovieUtilsSuccessMocked struct{}

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

func (mh MovieUtilsSuccessMocked) UpdateMovie(id int64, payload models.MovieUpdatePayload) (movie models.MovieDetail, err error) {
	movie.ID = id
	movie.Name = payload.MovieName
	movie.URL = payload.URL
	movie.DateOfLastWatchedEpisode, _ = time.Parse(time.RFC822Z, "29 Jan 91 03:04 -0700")
	movie.LastWatchedEpisode.ID = 2
	movie.LastWatchedEpisode.Series = 3
	movie.LastWatchedEpisode.EpisodeNumber = 3
	return movie, nil
}

func (mh MovieUtilsSuccessMocked) GetJSONParameters(body io.ReadCloser, out interface{}) error {
	return utils.GetJSONParameters(body, out)
}

func (mh MovieUtilsSuccessMocked) DeleteMovie(id int64) error {
	return nil
}

func (mh MovieUtilsSuccessMocked) RetrieveMovieDetail(movieID int64) (movie models.MovieDetail, err error) {
	movie.ID = 1
	movie.Name = "Test Movie 1"
	movie.URL = "http://www.example.com/movie1"
	movie.DateOfLastWatchedEpisode, _ = time.Parse(time.RFC822Z, "29 Jan 91 03:04 -0700")
	movie.LastWatchedEpisode.ID = 2
	movie.LastWatchedEpisode.Series = 3
	movie.LastWatchedEpisode.EpisodeNumber = 3
	movie.SeriesCount = 5
	return movie, nil
}
// MovieUtilsCreateFailedMocked
type MovieUtilsCreateFailedMocked struct{}

func (mh MovieUtilsCreateFailedMocked) CreateMovie(payload models.MovieCreationPayload) (movie models.MovieDetail, err error) {
	return movie, fmt.Errorf("Test error durring create movie")
}

func (mh MovieUtilsCreateFailedMocked) UpdateMovie(id int64, payload models.MovieUpdatePayload) (movie models.MovieDetail, err error) {
	return movie, nil
}

func (mh MovieUtilsCreateFailedMocked) GetJSONParameters(body io.ReadCloser, out interface{}) error {
	return utils.GetJSONParameters(body, out)
}

func (mh MovieUtilsCreateFailedMocked) DeleteMovie(id int64) error {
	return nil
}

func (mh MovieUtilsCreateFailedMocked) RetrieveMovieDetail(movieID int64) (movie models.MovieDetail, err error) {
	return movie, nil
}

// MovieUtilsJSONParseFailedMocked
type MovieUtilsJSONParseFailedMocked struct{}

func (mh MovieUtilsJSONParseFailedMocked) CreateMovie(payload models.MovieCreationPayload) (movie models.MovieDetail, err error) {
	return movie, fmt.Errorf("Test error durring create movie")
}

func (mh MovieUtilsJSONParseFailedMocked) UpdateMovie(id int64, payload models.MovieUpdatePayload) (movie models.MovieDetail, err error) {
	return movie, nil
}

func (mh MovieUtilsJSONParseFailedMocked) GetJSONParameters(body io.ReadCloser, out interface{}) error {
	return fmt.Errorf("Test error during parsing JSON parameters")
}

func (mh MovieUtilsJSONParseFailedMocked) DeleteMovie(id int64) error {
	return nil
}

func (mh MovieUtilsJSONParseFailedMocked) RetrieveMovieDetail(movieID int64) (movie models.MovieDetail, err error) {
	return movie, nil
}

// MovieUtilsUpdateMovieFailedMocked
type MovieUtilsUpdateMovieFailedMocked struct{}

func (mh MovieUtilsUpdateMovieFailedMocked) CreateMovie(payload models.MovieCreationPayload) (movie models.MovieDetail, err error) {
	return movie, nil
}

func (mh MovieUtilsUpdateMovieFailedMocked) UpdateMovie(id int64, payload models.MovieUpdatePayload) (movie models.MovieDetail, err error) {
	return movie, fmt.Errorf("Test error during update movie")
}

func (mh MovieUtilsUpdateMovieFailedMocked) GetJSONParameters(body io.ReadCloser, out interface{}) error {
	return nil
}

func (mh MovieUtilsUpdateMovieFailedMocked) DeleteMovie(id int64) error {
	return nil
}

func (mh MovieUtilsUpdateMovieFailedMocked) RetrieveMovieDetail(movieID int64) (movie models.MovieDetail, err error) {
	return movie, nil
}

// MovieUtilsDeleteMovieFailedMocked
type MovieUtilsDeleteMovieFailedMocked struct{}

func (mh MovieUtilsDeleteMovieFailedMocked) CreateMovie(payload models.MovieCreationPayload) (movie models.MovieDetail, err error) {
	return movie, nil
}

func (mh MovieUtilsDeleteMovieFailedMocked) UpdateMovie(id int64, payload models.MovieUpdatePayload) (movie models.MovieDetail, err error) {
	return movie, nil
}

func (mh MovieUtilsDeleteMovieFailedMocked) GetJSONParameters(body io.ReadCloser, out interface{}) error {
	return nil
}

func (mh MovieUtilsDeleteMovieFailedMocked) DeleteMovie(id int64) error {
	return fmt.Errorf("Test error during delete movie")
}

func (mh MovieUtilsDeleteMovieFailedMocked) RetrieveMovieDetail(movieID int64) (movie models.MovieDetail, err error) {
	movie.ID = 1
	return movie, nil
}


// MovieRetrieveDetailFailedMocked
type MovieRetrieveDetailFailedMocked struct{}

func (mh MovieRetrieveDetailFailedMocked) CreateMovie(payload models.MovieCreationPayload) (movie models.MovieDetail, err error) {
	return movie, nil
}

func (mh MovieRetrieveDetailFailedMocked) UpdateMovie(id int64, payload models.MovieUpdatePayload) (movie models.MovieDetail, err error) {
	return movie, nil
}

func (mh MovieRetrieveDetailFailedMocked) GetJSONParameters(body io.ReadCloser, out interface{}) error {
	return nil
}

func (mh MovieRetrieveDetailFailedMocked) DeleteMovie(id int64) error {
	return nil
}

func (mh MovieRetrieveDetailFailedMocked) RetrieveMovieDetail(movieID int64) (movie models.MovieDetail, err error) {
	return movie, fmt.Errorf("Test error during retrieve")
}
