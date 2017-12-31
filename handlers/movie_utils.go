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

		rows.Scan(&movie.ID, &movie.Name, &movie.URL)
		movies = append(movies, movie)
	}
	return movies, nil
}

// RetrieveMovieDetail found movie details
func (mh MovieHandlers) RetrieveMovieDetail(movieID int64) (movie models.MovieDetail, err error) {
	query := "SELECT tv_series.id, tv_series.name, url, COUNT(season.id) AS seriesCount FROM tv_series LEFT JOIN season ON season.serial_id = tv_series.id WHERE tv_series.id = ? GROUP BY tv_series.id;"
	rows, err := database.GetDBConn().Query(query, movieID)
	if err != nil {
		return movie, err
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&movie.ID, &movie.Name, &movie.URL, &movie.SeriesCount)

	query = "SELECT episode.id, season.id, episode.number, episode.date FROM episode JOIN season ON season.id = episode.season_id WHERE season.serial_id = ? AND episode.watched = 1 ORDER BY date DESC LIMIT 1;"
	rows, err = database.GetDBConn().Query(query, movieID)

	if err != nil {
		return movie, nil
	}

	defer rows.Close()
	rows.Next()
	rows.Scan(&movie.LastWatchedEpisode.ID, &movie.LastWatchedEpisode.Series, &movie.LastWatchedEpisode.EpisodeNumber, &movie.DateOfLastWatchedEpisode)

	return movie, nil
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
	return mh.RetrieveMovieDetail(movieID)
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
	return id, err
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

	return mh.RetrieveMovieDetail(movieID)
}

// GetJSONParameters function return params from ReadCloser object
func (mh MovieHandlers) GetJSONParameters(body io.ReadCloser, out interface{}) error {
	return utils.GetJSONParameters(body, out)
}

// DeleteMovie function execute delete query on database
func (mh MovieHandlers) DeleteMovie(movieID int64) error {
	conn := database.GetDBConn()

	stmt, err := conn.Prepare("DELETE FROM tv_series WHERE id = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(movieID)
	if err != nil {
		return err
	}

	return nil
}
