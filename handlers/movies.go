// Package movies provide an endpoint to operations on movie object
package movies

import (
	"io"
	"net/http"
	"strconv"

	"github.com/Mowinski/LastWatchedBackend/models"

	"github.com/gorilla/mux"

	"github.com/Mowinski/LastWatchedBackend/utils"
)

// Utils interface describe all utility function in handlers
type Utils interface {
	CreateMovie(payload models.MovieCreationPayload) (movie models.MovieDetail, err error)
	GetJSONParameters(body io.ReadCloser, out interface{}) error
}

// MovieHandlers join together all movie handlers
type MovieHandlers struct {
	Utils Utils
}

// MovieListHandler is responsive for return movie list
func (mh MovieHandlers) MovieListHandler(w http.ResponseWriter, r *http.Request) {
	searchString := "%" + r.URL.Query().Get("searchString") + "%"

	skip := utils.GetIntOrDefault(r.URL.Query().Get("skip"), 0)
	limit := utils.GetIntOrDefault(r.URL.Query().Get("limit"), 50)

	movies, err := retriveMovieItems(searchString, limit, skip)
	if err != nil {
		utils.ResponseBadRequestError(w, err)
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, movies)
}

// MovieDetailsHanlder is responsive for return movie detials
func (mh MovieHandlers) MovieDetailsHanlder(w http.ResponseWriter, r *http.Request) {
	movieID, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)

	movie, err := retriveMovieDetail(movieID)
	if err != nil {
		utils.ResponseBadRequestError(w, err)
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, movie)
}

// MovieCreateHandler create new movie in database
func (mh MovieHandlers) MovieCreateHandler(w http.ResponseWriter, r *http.Request) {
	var payload models.MovieCreationPayload
	err := mh.Utils.GetJSONParameters(r.Body, &payload)
	if err != nil {
		utils.ResponseBadRequestError(w, err)
		return
	}

	movie, err := mh.Utils.CreateMovie(payload)
	if err != nil {
		utils.ResponseBadRequestError(w, err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, movie)
}
