// Package movies provide an endpoint to operations on movie object
package movies

import (
	"fmt"
	"net/http"
)

// IndexHandler is responsive for return movie list
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
}
