package httpServer

import (
	"expvar"
	"github.com/gorilla/mux"
	"github.com/lestrrat-go/apache-logformat"
	"io"
	"net/http"
)

type HttpRoute struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc HandlerHandleFunc
}

var httpRoutes = []HttpRoute{
	{
		"StatsCounts",
		"GET",
		"/api/v0/Stats/Counts",
		HandleStatsCounts,
	}, {
		"ApiIndex",
		"GET",
		"/api{Path:.*}",
		HandleApiNotFound,
	}, {
		"expvar",
		"GET",
		"/debug/vars",
		func(env *Environment, w http.ResponseWriter, r *http.Request) Error {
			expvar.Handler().ServeHTTP(w, r)
			return nil
		},
	},
}

func newRouter(logger io.Writer, env *Environment) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// setup normal http routes
	for _, route := range httpRoutes {
		var handler http.Handler
		handler = Handler{Env: env, Handle: route.HandlerFunc}
		if logger != nil {
			handler = apachelog.CombinedLog.Wrap(handler, logger)
		}

		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}

	return router
}
