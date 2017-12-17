// Package movies provide an endpoint to operations on movie object
package movies

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Mowinski/LastWatchedBackend/logger"
	"github.com/Mowinski/LastWatchedBackend/models"
	"github.com/Mowinski/LastWatchedBackend/utils"

	"github.com/Mowinski/LastWatchedBackend/database"
)

// MovieListHandler is responsive for return movie list
func MovieListHandler(w http.ResponseWriter, r *http.Request) {
	searchString := "%" + r.URL.Query().Get("searchString") + "%"

	skip := utils.GetIntOrDefault(r.URL.Query().Get("skip"), 0)
	limit := utils.GetIntOrDefault(r.URL.Query().Get("limit"), 50)

	rows, err := database.GetDBConn().Query("SELECT id, name, url  FROM tv_series WHERE name LIKE ? LIMIT ? OFFSET ?;", searchString, limit, skip)
	if err != nil {
		logger.Logger.Print("Error during fetch movies from database, error: ", err)
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]error{"error": err})
		return
	}
	defer rows.Close()

	movies := retriveMovieItems(rows)

	utils.RespondWithJSON(w, http.StatusOK, movies)
}

func retriveMovieItems(rows *sql.Rows) (movies models.MovieItems) {
	for rows.Next() {
		var movie models.MovieItem

		if err := rows.Scan(&movie.ID, &movie.Name, &movie.URL); err != nil {
			log.Print("Error during scan data from row, error: ", err)
		}
		movies = append(movies, movie)
	}
	return movies
}
