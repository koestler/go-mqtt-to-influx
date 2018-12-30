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

func HandleApiNotFound(env *Environment, w http.ResponseWriter, r *http.Request) Error {
	err := errors.New("api method not found")
	return StatusError{404, err}
}

func HandleStatsCounts(env *Environment, w http.ResponseWriter, r *http.Request) Error {
	if env.Statistics == nil {
		// Statistics module not available -> return 404
		err := errors.New("Statistics module is disabled")
		return StatusError{404, err}
	}

	counts := env.Statistics.GetTotalPerModule()

	writeJsonHeaders(w)
	b, err := json.MarshalIndent(counts, "", "    ")
	if err != nil {
		return StatusError{500, err}
	}
	w.Write(b)
	return nil
}
