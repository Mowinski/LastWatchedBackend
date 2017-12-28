package movies

import (
	"database/sql"

	"io"

	"github.com/Mowinski/LastWatchedBackend/database"
	"github.com/Mowinski/LastWatchedBackend/models"
	"github.com/Mowinski/LastWatchedBackend/utils"
)

func retrieveMovieItems(searchString string, limit int, skip int) (movies models.MovieItems, err error) {
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

func retrieveMovieDetail(movieID int64) (movie models.MovieDetail, err error) {
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
		return movie, err
	}

	return movie, err
}

// CreateMovie function create movie in database
func (mh MovieHandlers) CreateMovie(payload models.MovieCreationPayload) (movie models.MovieDetail, err error) {
	conn := database.GetDBConn()

	tx, err := conn.Begin()
	if err != nil {
		return movie, err
	}

	movieID, err := executeStmt(
		tx,
		"INSERT INTO tv_series (name, url) VALUES (?, ?);",
		payload.MovieName,
		payload.URL,
	)

	if err != nil {
		return movie, err
	}

	for seriesNumber := 1; seriesNumber <= payload.SeriesNumber; seriesNumber++ {
		seriesID, err := executeStmt(tx, "INSERT INTO season (serial_id, number) VALUES (?, ?)", movieID, seriesNumber)
		if err != nil {
			tx.Rollback()
			return movie, err
		}
		for episodeNumber := 1; episodeNumber <= payload.EpisodesInSeries; episodeNumber++ {
			_, err := executeStmt(
				tx,
				"INSERT INTO episode (season_id, number, watched, date) VALUES (?, ?, 0, null);",
				seriesID,
				episodeNumber,
			)
			if err != nil {
				tx.Rollback()
				return movie, err
			}
		}
	}
	tx.Commit()
	return retrieveMovieDetail(movieID)
}

func executeStmt(tx *sql.Tx, query string, args ...interface{}) (id int64, err error) {
	stmt, err := tx.Prepare(query)
	if err != nil {
		return id, err
	}

	rows, err := stmt.Exec(args...)
	if err != nil {
		return id, err
	}

	id, err = rows.LastInsertId()
	if err != nil {
		return id, err
	}
	return id, nil
}

// UpdateMovie function update selected movie
func (mh MovieHandlers) UpdateMovie(movieID int64, payload models.MovieUpdatePayload) (movie models.MovieDetail, err error) {
	conn := database.GetDBConn()

	tx, err := conn.Begin()
	if err != nil {
		return movie, err
	}

	_, err = executeStmt(
		tx,
		"UPDATE tv_series SET name = ?, url = ? WHERE id = ?;",
		payload.MovieName,
		payload.URL,
		movieID,
	)

	if err != nil {
		return movie, err
	}

	tx.Commit()

	return retrieveMovieDetail(movieID)
}

// GetJSONParameters function return params from ReadCloser object
func (mh MovieHandlers) GetJSONParameters(body io.ReadCloser, out interface{}) error {
	return utils.GetJSONParameters(body, out)
}
