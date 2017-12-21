package main

import (
	"net/http"
	"time"

	"github.com/Mowinski/LastWatchedBackend/handlers"
	"github.com/Mowinski/LastWatchedBackend/logger"
	"github.com/gorilla/mux"
)

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

func newRouter() *mux.Router {
	var movieHandler movies.MovieHandlers
	routes := []route{
		route{"MovieList", "GET", "/movies", movieHandler.MovieListHandler},
		route{"MovieDetail", "GET", "/movie/{id:[0-9]+}", movieHandler.MovieDetailsHanlder},
		route{"MovieCreate", "POST", "/movie", movieHandler.MovieCreateHandler},
	}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		handler := loggerHandler(route.HandlerFunc, route.Name)
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func loggerHandler(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		logger.Logger.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}
