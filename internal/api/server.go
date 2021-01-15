package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/tmnhat2001/worker-service/internal/worker"
	"golang.org/x/crypto/bcrypt"
)

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

	router.Handle("/start", server.makeHandler(server.startJob)).Methods("POST")
	router.Handle("/stop", server.makeHandler(server.stopJob)).Methods("PUT")
	router.Handle("/jobs/{jobID}", server.makeHandler(server.getJobResults)).Methods("GET")

	return router
}

func (server *Server) makeHandler(fn func(http.ResponseWriter, *http.Request)) http.Handler {
	return server.authHandler(http.HandlerFunc(fn))
}

func (server *Server) authHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}

func (server *Server) close() {
	err := server.httpServer.Close()
	if err != nil {
		log.Println(err)
	}
}

func (server *Server) startJob(w http.ResponseWriter, req *http.Request) {
	user, ok := userFromContext(req.Context())
	if !ok {
		server.logger.WithFields(logrus.Fields{
			"endpoint": req.URL.Path,
		}).Error("Unable to retrieve authenticated user")

		errorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(req.Body)

	var job worker.Job
	err := decoder.Decode(&job)
	if err != nil {
		server.logger.WithFields(logrus.Fields{
			"endpoint": req.URL.Path,
			"user":     user.Username,
		}).Error(err)

		errorResponse(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	service := jobService{jobStore: server.jobStore, user: user}
	updatedJob, err := service.startJob(&job)
	if err != nil {
		server.logger.WithFields(logrus.Fields{
			"endpoint": req.URL.Path,
			"user":     user.Username,
		}).Error(err)

		errorResponse(w, "Failed to start job", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, updatedJob, http.StatusOK)
}

func (server *Server) stopJob(w http.ResponseWriter, req *http.Request) {
	user, ok := userFromContext(req.Context())
	if !ok {
		server.logger.WithFields(logrus.Fields{
			"endpoint": req.URL.Path,
		}).Error("Unable to retrieve authenticated user")

		errorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(req.Body)

	var jobRequest worker.Job
	err := decoder.Decode(&jobRequest)
	if err != nil {
		server.logger.WithFields(logrus.Fields{
			"endpoint": req.URL.Path,
			"user":     user.Username,
		}).Error(err)

		errorResponse(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	service := jobService{jobStore: server.jobStore, user: user}
	job, err := service.stopJob(jobRequest.ID)
	if (err == errUnauthorizedUser) || (err == worker.ErrJobNotFound) {
		server.logger.WithFields(logrus.Fields{
			"endpoint": req.URL.Path,
			"user":     user.Username,
		}).Error(err)

		errorResponse(w, "Failed to find job", http.StatusNotFound)
		return
	} else if err != nil {
		server.logger.WithFields(logrus.Fields{
			"endpoint": req.URL.Path,
			"user":     user.Username,
		}).Error(err)

		errorResponse(w, "Failed to stop job. The job may have already finished.", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, job, http.StatusOK)
}

func (server *Server) getJobResults(w http.ResponseWriter, req *http.Request) {
	user, ok := userFromContext(req.Context())
	if !ok {
		server.logger.WithFields(logrus.Fields{
			"endpoint": req.URL.Path,
		}).Error("Unable to retrieve authenticated user")

		errorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	requestVars := mux.Vars(req)
	service := jobService{jobStore: server.jobStore, user: user}
	job, err := service.getJob(requestVars["jobID"])
	if (err == errUnauthorizedUser) || (err == worker.ErrJobNotFound) {
		server.logger.WithFields(logrus.Fields{
			"endpoint": req.URL.Path,
			"user":     user.Username,
		}).Error(err)

		errorResponse(w, "Failed to find job", http.StatusNotFound)
		return
	} else if err != nil {
		server.logger.WithFields(logrus.Fields{
			"endpoint": req.URL.Path,
			"user":     user.Username,
		}).Error(err)

		errorResponse(w, "An unexpected error has occurred", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, job, http.StatusOK)
}
