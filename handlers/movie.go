// Package movies provide an endpoint to operations on movie object
package movies

import (
	"net/http"
	"strconv"

	"github.com/Mowinski/LastWatchedBackend/models"

	"github.com/gorilla/mux"

	"github.com/Mowinski/LastWatchedBackend/utils"
)

// MovieListHandler is responsive for return movie list
func MovieListHandler(w http.ResponseWriter, r *http.Request) {
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
func MovieDetailsHanlder(w http.ResponseWriter, r *http.Request) {
	movieID, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)

	movie, err := retriveMovieDetail(movieID)
	if err != nil {
		utils.ResponseBadRequestError(w, err)
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, movie)
}

// MovieCreateHandler create new movie in database
func MovieCreateHandler(w http.ResponseWriter, r *http.Request) {
	var payload models.MovieCreationPayload
	err := utils.GetJSONParameters(r.Body, &payload)
	if err != nil {
		utils.ResponseBadRequestError(w, err)
		return
	}

	movie, err := createMovie(payload)
	if err != nil {
		utils.ResponseBadRequestError(w, err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, movie)
}
