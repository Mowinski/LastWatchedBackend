package models

// MovieItem is stuct which contains simple information about movie
type MovieItem struct {
	ID   int
	Name string
	URL  string
}

// MovieItems is array type which contains list of MovieItems
type MovieItems []MovieItem
