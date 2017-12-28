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

	movieCreatePayload           movieBodyPayload
	movieUpdatePayload           movieBodyPayload
	movieSuccessHandlers         movies.MovieHandlers
	movieCreateFailedHandlers    movies.MovieHandlers
	movieJSONParseFailedHandlers movies.MovieHandlers
	movieUpdateFailedHandlers    movies.MovieHandlers
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
