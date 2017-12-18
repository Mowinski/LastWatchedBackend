package movies

import (
	"strconv"

	"github.com/Mowinski/LastWatchedBackend/database"
	"github.com/Mowinski/LastWatchedBackend/models"
)

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

func retriveMovieDetail(movieID int64) (movie models.MovieDetail, err error) {
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

	query = "SELECT episode.id, season.id, episode.number, episode.date FROM episode JOIN season ON season.id = episode.season_id WHERE season.serial_id = ? AND episode.watched = 1 ORDER BY date DESC LIMIT 1;"
	rows, err = database.GetDBConn().Query(query, movieID)
	if err != nil {
		return movie, err
	}

	defer rows.Close()
	rows.Next()
	err = rows.Scan(&movie.LastWatchedEpisode.ID, &movie.LastWatchedEpisode.Series, &movie.LastWatchedEpisode.EpisodeNumber, &movie.DateOfLastWatchedEpisode)
	if err != nil {
		return movie, nil
	}

	return movie, err
}

func createMovie(payload models.MovieCreationPayload) (movie models.MovieDetail, err error) {
	conn := database.GetDBConn()

	tx, err := conn.Begin()
	if err != nil {
		return movie, err
	}

	stmt, err := conn.Prepare("INSERT INTO tv_series (name, url) VALUES (?, ?);")
	if err != nil {
		return movie, err
	}
	rows, err := stmt.Exec(payload.MovieName, payload.URL)
	if err != nil {
		return movie, err
	}

	movieID, err := rows.LastInsertId()
	if err != nil {
		return movie, err
	}

	for i := 1; i <= payload.SeriesNumber; i++ {
		_, err := conn.Query("INSERT INTO season (serial_id, name) VALUES (?, ?)", movieID, "Sezon "+strconv.Itoa(i))
		if err != nil {
			tx.Rollback()
			return movie, err
		}
	}
	tx.Commit()
	return retriveMovieDetail(movieID)
}
