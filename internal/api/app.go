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

const hostname = "localhost"

// App represents an object that handles API requests
type App struct {
	jobStore    worker.JobStore
	authService *AuthenticationService
	server      *http.Server
}

// RunApp creates an instance of the App and runs it
func RunApp() {
	app := NewApp(8080)
	app.run("certs/server.crt", "certs/server.key")
}

// NewApp returns a new App instance
func NewApp(port int) *App {
	users, err := createUsers()
	if err != nil {
		log.Fatal(err)
	}

	userRepository := &MemoryUserRepository{Users: users}
	jobs := make(map[string]worker.Job)
	app := &App{
		jobStore:    &worker.MemoryJobStore{Jobs: jobs},
		authService: &AuthenticationService{UserRepository: userRepository},
	}

	router := app.registerRoutes()
	address := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:    address,
		Handler: router,
	}
	app.server = server

	return app
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

func (app *App) run(certFilePath, keyFilePath string) {
	err := app.server.ListenAndServeTLS(certFilePath, keyFilePath)
	log.Println(err)
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

	job.User = user.Username
	err = (&job).Start(app.jobStore)
	if err != nil {
		log.Println(err)
		errorResponse(w, "Failed to start job", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, job, http.StatusOK)
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

	job, err := app.jobStore.FindJob(jobRequest.ID)
	if err != nil {
		log.Println(err)
		errorResponse(w, "Failed to find job", http.StatusNotFound)
		return
	}

	if job.User != user.Username {
		errorResponse(w, "Failed to find job", http.StatusNotFound)
		return
	}

	err = job.Stop(app.jobStore)
	if err != nil {
		log.Println(err)
		errorResponse(w, "Failed to stop job. The job may have already finished.", http.StatusInternalServerError)
		return
	}

	updatedJob, err := app.jobStore.FindJob(jobRequest.ID)
	if err != nil {
		log.Println(err)
		errorResponse(w, "Failed to find job", http.StatusNotFound)
		return
	}

	jsonResponse(w, updatedJob, http.StatusOK)
}

func (app *App) getJobResults(w http.ResponseWriter, req *http.Request) {
	requestVars := mux.Vars(req)
	job, err := app.jobStore.FindJob(requestVars["jobID"])
	if err != nil {
		log.Println(err)
		errorResponse(w, "Failed to find job", http.StatusNotFound)
		return
	}

	user, ok := userFromContext(req.Context())
	if !ok {
		log.Println("Unable to retrieve authenticated user")
		errorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if job.User != user.Username {
		errorResponse(w, "Failed to find job", http.StatusNotFound)
		return
	}

	jsonResponse(w, job, http.StatusOK)
}
