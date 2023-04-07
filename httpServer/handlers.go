package httpServer

import (
	"encoding/json"
	"errors"
	"net/http"
)

func writeJsonHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Model", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func HandleStatsCounts(env *Environment, w http.ResponseWriter, r *http.Request) Error {
	if !env.Statistics.Enabled() {
		// Statistics module not available -> return 404
		err := errors.New("statistics module is disabled")
		return StatusError{404, err}
	}

	counts := env.Statistics.GetHierarchicalCountsStructless()

	writeJsonHeaders(w)
	b, err := json.MarshalIndent(counts, "", "    ")
	if err != nil {
		return StatusError{500, err}
	}
	_, err = w.Write(b)
	if err != nil {
		return StatusError{500, err}
	}
	return nil
}
