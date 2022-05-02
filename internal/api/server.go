package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/MarySmirnova/api_users/internal/config"
	"github.com/MarySmirnova/api_users/internal/database"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

var ErrPermissionsDenied error = errors.New("insufficient permissions")

type Storage interface {
	NewUser(*database.User) error
	GetAllUsers() []*database.User
	GetUserByID(uuid.UUID) (*database.User, error)
	GetUserByName(string) (*database.User, error)
	UpdateUser(*database.User) error
	DeleteUser(uuid.UUID) error
}

type API struct {
	store      Storage
	httpServer *http.Server
}

func New(cfg config.API, s Storage) *API {
	a := &API{
		store: s,
	}

	handler := mux.NewRouter()
	handler.Use(a.AuthMiddleware)
	handler.Name("create_user").Methods(http.MethodPost).Path("/user").HandlerFunc(a.NewUserHandler)
	handler.Name("get_all_users").Methods(http.MethodGet).Path("/user").HandlerFunc(a.GetUsersHandler)
	handler.Name("get_user").Methods(http.MethodGet).Path("/user/{id}").HandlerFunc(a.GetUserByIDHandler)
	handler.Name("update_user").Methods(http.MethodPatch).Path("/user/{id}").HandlerFunc(a.UpdateUserHandler)
	handler.Name("delete_user").Methods(http.MethodDelete).Path("/user/{id}").HandlerFunc(a.DeleteUserHandler)

	a.httpServer = &http.Server{
		Addr:         cfg.Listen,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return a
}

func (a *API) GetHTTPServer() *http.Server {
	return a.httpServer
}

func (a *API) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			a.askPassword(w)
			return
		}

		user, err := a.store.GetUserByName(username)
		if err != nil {
			if errors.Is(err, database.ErrUserNotExist) {
				a.askPassword(w)
				return
			}
			a.internalError(w, err)
			return
		}

		if !user.CheckPassword(password) {
			a.askPassword(w)
			return
		}

		ctx := context.WithValue(r.Context(), ContextAdminKey, user.Admin)

		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) askPassword(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	w.WriteHeader(http.StatusUnauthorized)
}

func (a *API) internalError(w http.ResponseWriter, err error) {
	log.WithError(err).Error("unable to get user from the store")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte("something went wrong"))
}

func (a *API) writeResponseError(w http.ResponseWriter, err error, code int) {
	log.WithError(err).Error("api error")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(err.Error()))
}
