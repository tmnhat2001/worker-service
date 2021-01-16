package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tmnhat2001/worker-service/internal/worker"
	"golang.org/x/crypto/bcrypt"
)

type customHandler func(r *http.Request) (worker.Job, requestError)

// Server represents server that handles API requests
type Server struct {
	jobStore    worker.JobStore
	authService *AuthenticationService
	httpServer  *http.Server
	logger      *logrus.Logger
	config      ServerConfig
}

// NewServer returns a new Server instance
func NewServer(config ServerConfig) (*Server, error) {
	authService, err := newAuthenticationService()
	if err != nil {
		return nil, err
	}

	server := &Server{
		jobStore: &worker.MemoryJobStore{
			Jobs: make(map[string]worker.Job),
		},
		authService: authService,
		logger:      logrus.New(),
		config:      config,
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: server.registerRoutes(),
	}
	server.httpServer = httpServer

	return server, nil
}

// Run starts the Server
func (server *Server) Run() error {
	return server.httpServer.ListenAndServeTLS(server.config.CertFilePath, server.config.KeyFilePath)
}

func (server *Server) registerRoutes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/start", server.makeHandler(server.startJob)).Methods("POST")
	router.Handle("/stop", server.makeHandler(server.stopJob)).Methods("PUT")
	router.Handle("/jobs/{jobID}", server.makeHandler(server.getJobResults)).Methods("GET")

	return router
}

func (server *Server) makeHandler(fn customHandler) http.HandlerFunc {
	return server.authHandler(server.requestHandler(fn))
}

func (server *Server) authHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := server.authService.Authenticate(r)
		if err != nil {
			if err != bcrypt.ErrMismatchedHashAndPassword {
				server.logger.WithFields(logrus.Fields{"endpoint": r.URL.Path}).Error(err)
			}

			errorResponse(w, "Unable to authenticate user", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		key := contextKey("user")
		newCtx := context.WithValue(ctx, key, user)
		newRequest := r.WithContext(newCtx)
		next.ServeHTTP(w, newRequest)
	}
}

func (server *Server) requestHandler(fn customHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		job, err := fn(req)
		if (err != requestError{}) {
			server.logger.WithFields(logrus.Fields{
				"endpoint": req.URL.Path,
			}).Error(errors.Unwrap(err))

			errorResponse(w, err.message, err.statusCode)
			return
		}

		jsonResponse(w, job, http.StatusOK)
	}
}

func (server *Server) close() {
	err := server.httpServer.Close()
	if err != nil {
		server.logger.Error(err)
	}
}

func (server *Server) startJob(req *http.Request) (worker.Job, requestError) {
	user, err := userFromContext(req.Context())
	if err != nil {
		return worker.Job{}, requestError{wrappedError: err, message: "Internal server error", statusCode: http.StatusInternalServerError}
	}

	var job worker.Job
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&job)
	if err != nil {
		return worker.Job{}, requestError{wrappedError: err, message: "Failed to parse request", statusCode: http.StatusNotFound}
	}

	service := jobService{jobStore: server.jobStore, user: user}
	updatedJob, err := service.startJob(&job)
	if err != nil {
		return worker.Job{}, requestError{wrappedError: err, message: "Failed to start job", statusCode: http.StatusInternalServerError}
	}

	return updatedJob, requestError{}
}

func (server *Server) stopJob(req *http.Request) (worker.Job, requestError) {
	user, err := userFromContext(req.Context())
	if err != nil {
		return worker.Job{}, requestError{wrappedError: err, message: "Internal server error", statusCode: http.StatusInternalServerError}
	}

	var jobRequest worker.Job
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&jobRequest)
	if err != nil {
		return worker.Job{}, requestError{wrappedError: err, message: "Failed to parse request", statusCode: http.StatusNotFound}
	}

	service := jobService{jobStore: server.jobStore, user: user}
	job, err := service.stopJob(jobRequest.ID)
	if (err == errUnauthorizedUser) || (err == worker.ErrJobNotFound) {
		return worker.Job{}, requestError{wrappedError: err, message: "Failed to find job", statusCode: http.StatusNotFound}
	} else if err != nil {
		return worker.Job{}, requestError{wrappedError: err, message: "Failed to stop job. The job may have already finished.", statusCode: http.StatusInternalServerError}
	}

	return job, requestError{}
}

func (server *Server) getJobResults(req *http.Request) (worker.Job, requestError) {
	user, err := userFromContext(req.Context())
	if err != nil {
		return worker.Job{}, requestError{wrappedError: err, message: "Internal server error", statusCode: http.StatusInternalServerError}
	}

	requestVars := mux.Vars(req)
	service := jobService{jobStore: server.jobStore, user: user}
	job, err := service.getJob(requestVars["jobID"])
	if (err == errUnauthorizedUser) || (err == worker.ErrJobNotFound) {
		return worker.Job{}, requestError{wrappedError: err, message: "Failed to find job", statusCode: http.StatusNotFound}
	} else if err != nil {
		return worker.Job{}, requestError{wrappedError: err, message: "An unexpected error has occurred", statusCode: http.StatusInternalServerError}
	}

	return job, requestError{}
}
