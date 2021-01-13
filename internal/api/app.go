package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tmnhat2001/worker-service/internal/worker"
	"golang.org/x/crypto/bcrypt"
)

// App represents server that handles API requests
type App struct {
	jobStore    worker.JobStore
	authService *AuthenticationService
	server      *http.Server
}

// RunApp creates an instance of the App and runs it
func RunApp() error {
	app, err := NewApp(8080)
	if err != nil {
		return err
	}

	err = app.run("certs/server.crt", "certs/server.key")
	return err
}

// NewApp returns a new App instance
func NewApp(port int) (*App, error) {
	users, err := createUsers()
	if err != nil {
		return nil, err
	}

	app := &App{
		jobStore: &worker.MemoryJobStore{
			Jobs: make(map[string]worker.Job),
		},
		authService: &AuthenticationService{
			UserRepository: &MemoryUserRepository{Users: users},
		},
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: app.registerRoutes(),
	}
	app.server = server

	return app, nil
}

func (app *App) registerRoutes() *mux.Router {
	router := mux.NewRouter()

	router.Handle("/start", app.makeHandler(app.startJob)).Methods("POST")
	router.Handle("/stop", app.makeHandler(app.stopJob)).Methods("PUT")
	router.Handle("/jobs/{jobID}", app.makeHandler(app.getJobResults)).Methods("GET")

	return router
}

func (app *App) makeHandler(fn func(http.ResponseWriter, *http.Request)) http.Handler {
	return app.authHandler(http.HandlerFunc(fn))
}

func (app *App) authHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := app.authService.Authenticate(r)
		if err != nil {
			if err != bcrypt.ErrMismatchedHashAndPassword {
				log.Println(err)
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

func (app *App) run(certFilePath, keyFilePath string) error {
	err := app.server.ListenAndServeTLS(certFilePath, keyFilePath)
	return err
}

func (app *App) close() {
	err := app.server.Close()
	if err != nil {
		log.Println(err)
	}
}

func (app *App) startJob(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)

	var job worker.Job
	err := decoder.Decode(&job)
	if err != nil {
		log.Println(err)
		errorResponse(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	user, ok := userFromContext(req.Context())
	if !ok {
		log.Println("Unable to retrieve authenticated user")
		errorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	service := jobService{jobStore: app.jobStore, user: user}
	updatedJob, err := service.startJob(&job)
	if err != nil {
		log.Println(err)
		errorResponse(w, "Failed to start job", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, updatedJob, http.StatusOK)
}

func (app *App) stopJob(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)

	var jobRequest worker.Job
	err := decoder.Decode(&jobRequest)
	if err != nil {
		log.Println(err)
		errorResponse(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	user, ok := userFromContext(req.Context())
	if !ok {
		log.Println("Unable to retrieve authenticated user")
		errorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	service := jobService{jobStore: app.jobStore, user: user}
	job, err := service.stopJob(jobRequest.ID)
	if (err == errUnauthorizedUser) || (err == worker.ErrJobNotFound) {
		log.Println(err)
		errorResponse(w, "Failed to find job", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println(err)
		errorResponse(w, "Failed to stop job. The job may have already finished.", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, job, http.StatusOK)
}

func (app *App) getJobResults(w http.ResponseWriter, req *http.Request) {
	user, ok := userFromContext(req.Context())
	if !ok {
		log.Println("Unable to retrieve authenticated user")
		errorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	requestVars := mux.Vars(req)
	service := jobService{jobStore: app.jobStore, user: user}
	job, err := service.getJob(requestVars["jobID"])
	if (err == errUnauthorizedUser) || (err == worker.ErrJobNotFound) {
		log.Println(err)
		errorResponse(w, "Failed to find job", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println(err)
		errorResponse(w, "An unexpected error has occurred", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, job, http.StatusOK)
}
