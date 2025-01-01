package web_errors

import (
	"fmt"
	"net/http"
)

func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "<p>404</p>")
		http.Redirect(w, r, "/", 404)
	}
}
