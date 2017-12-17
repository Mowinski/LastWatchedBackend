// Package movies provide an endpoint to operations on movie object
package movies

import (
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

	movies, err := retriveMovieItems(searchString, limit, skip)
	if err != nil {
		logger.Logger.Print("Error during fetch movies from database, error: ", err)
		utils.RespondWithJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": err.Error()},
		)
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, movies)
}

func retriveMovieItems(searchString string, limit int, skip int) (movies models.MovieItems, err error) {
	rows, err := database.GetDBConn().Query("SELECT id, name, url  FROM tv_series WHERE name LIKE ? LIMIT ? OFFSET ?;", searchString, limit, skip)
	if err != nil {
		return movies, err
	}
	defer rows.Close()

	for rows.Next() {
		var movie models.MovieItem

		if err := rows.Scan(&movie.ID, &movie.Name, &movie.URL); err != nil {
			log.Print("Error during scan data from row, error: ", err)
		}
		movies = append(movies, movie)
	}
	return movies, nil
}
