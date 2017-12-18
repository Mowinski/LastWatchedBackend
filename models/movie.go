package models

import "time"

// MovieItem is stuct which contains simple information about movie
type MovieItem struct {
	ID   int
	Name string
	URL  string
}

// MovieItems is array type which contains list of MovieItems
type MovieItems []MovieItem

// MovieDetail descrbie details about selected movie series
type MovieDetail struct {
	ID                       int
	Name                     string
	URL                      string
	SeriesCount              int
	LastWatchedEpisode       Episode
	DateOfLastWatchedEpisode time.Time
}

// Episode describe one episode
type Episode struct {
	ID            int
	Series        int
	EpisodeNumber int
}
