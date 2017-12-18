// Package movies provide an endpoint to operations on movie object
package movies

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

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

// MovieDetailsHanlder is responsive for return movie detials
func MovieDetailsHanlder(w http.ResponseWriter, r *http.Request) {
	movieID, _ := strconv.Atoi(mux.Vars(r)["id"])

	movie, err := retriveMovieDetail(movieID)
	if err != nil {
		logger.Logger.Print("Error during fetch movies from database, error: ", err)
		utils.RespondWithJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": err.Error()},
		)
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, movie)
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
			return movies, err
		}
		movies = append(movies, movie)
	}
	return movies, nil
}

func retriveMovieDetail(movieID int) (movie models.MovieDetail, err error) {
	query := "SELECT tv_series.id, tv_series.name, url, COUNT(season.id) AS seriesCount FROM tv_series LEFT JOIN season ON season.serial_id = tv_series.id WHERE tv_series.id = ? GROUP BY tv_series.id;"
	rows, err := database.GetDBConn().Query(query, movieID)
	if err != nil {
		return movie, err
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&movie.ID, &movie.Name, &movie.URL, &movie.SeriesCount)
	if err != nil {
		return movie, err
	}

	return movie, err
}
