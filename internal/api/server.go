package api

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/MarySmirnova/api_users/internal/config"
	"github.com/MarySmirnova/api_users/internal/database"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ctxKey string

const isAdmin ctxKey = "is_admin"

type Storage interface {
	NewUser(database.User) (uuid.UUID, error)
	GetAllUsers() []database.User
	GetUserByID(uuid.UUID) (database.User, error)
	GetUserByName(string) (database.User, error)
	UpdateUser(database.User) error
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
	handler.Name("create_user").Methods(http.MethodPost).Path("/user").HandlerFunc(a.newUserHandler)
	handler.Name("get_all_users").Methods(http.MethodGet).Path("/user").HandlerFunc(a.getUsersHandler)
	handler.Name("get_user").Methods(http.MethodGet).Path("/user/{id}").HandlerFunc(a.getUserByIDHandler)
	handler.Name("update_user").Methods(http.MethodPatch).Path("/user/{id}").HandlerFunc(a.updateUserHandler)
	handler.Name("delete_user").Methods(http.MethodDelete).Path("/user/{id}").HandlerFunc(a.deleteUserHandler)

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
		uname, pass, ok := r.BasicAuth()
		if !ok {
			u, err := a.store.GetUserByName(uname)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			passHash := sha256.Sum256([]byte(pass))
			expectPassHash := sha256.Sum256([]byte(u.Password))
			matchPass := (subtle.ConstantTimeCompare(passHash[:], expectPassHash[:]) == 1)

			permis := u.Admin

			if matchPass {
				ctx := context.WithValue(r.Context(), isAdmin, permis)
				w.Header().Set("Content-Type", "application/json")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		callError(w, ErrUnauthorized, http.StatusUnauthorized)
	})
}
