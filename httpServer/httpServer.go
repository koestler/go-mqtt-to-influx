package httpServer

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type HttpServer struct {
	config Config
	server *http.Server
}

type Config interface {
	Bind() string
	Port() int
	LogRequests() bool
}

type Statistics interface {
	Enabled() bool
	GetHierarchicalCounts() interface{}
}

func Run(config Config, env *Environment) (httpServer *HttpServer) {
	var logger io.Writer
	if config.LogRequests() {
		logger = os.Stdout
	}

	address := config.Bind() + ":" + strconv.Itoa(config.Port())
	router := newRouter(logger, env)

	server := &http.Server{
		Addr:    address,
		Handler: router,
	}

	go func() {
		log.Printf("httpServer: listening on %v", address)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("httpServer: stopped due to error: %s", err)
		}
	}()

	return &HttpServer{
		config: config,
		server: server,
	}
}

func (s *HttpServer) Shutdown() {
	err := s.server.Shutdown(nil)
	if err != nil {
		log.Printf("httpServer: gracefully shutdown failed: %s", err)
	}
}

// Our application wide data containers
type Environment struct {
	Statistics Statistics
}

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type StatusError struct {
	Code int
	Err  error
}

// Allows StatusError to satisfy the error interface.
func (statusError StatusError) Error() string {
	return statusError.Err.Error()
}

// Returns our HTTP status code.
func (statusError StatusError) Status() int {
	return statusError.Code
}

// define an extended version of http.HandlerFunc
type HandlerHandleFunc func(e *Environment, w http.ResponseWriter, r *http.Request) Error

// The Handler struct that takes a configured Environment and a function matching
// our useful signature.
type Handler struct {
	Env    *Environment
	Handle HandlerHandleFunc
}

// ServeHTTP allows our Handler type to satisfy httpServer.Handler.
func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := handler.Handle(handler.Env, w, r)

	if err != nil {
		log.Printf("ServeHTTP err=%v", err)

		switch e := err.(type) {
		case Error:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			log.Printf("HTTP %d - %s", e.Status(), e)
			http.Error(w, http.StatusText(e.Status()), e.Status())
			return
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
	}
}
