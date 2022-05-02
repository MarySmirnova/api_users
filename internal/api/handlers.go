package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MarySmirnova/api_users/internal/database"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var validate = validator.New()

func (a *API) NewUserHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminUser(r.Context()) {
		a.writeResponseError(w, ErrPermissionsDenied, http.StatusForbidden)
		return
	}

	var u database.User

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		a.writeResponseError(w, fmt.Errorf("wrong JSON: %s", err), http.StatusBadRequest)
		return
	}

	if err := validate.Struct(u); err != nil {
		a.writeResponseError(w, fmt.Errorf("invalid data passed: %s", err), http.StatusBadRequest)
		return
	}

	if err := a.store.NewUser(&u); err != nil {
		if errors.Is(err, database.ErrNameAlreadyExist) {
			a.writeResponseError(w, err, http.StatusBadRequest)
			return
		}
		a.internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(u.ID)
}

func (a *API) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	users := a.store.GetAllUsers()

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(users)
}

func (a *API) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	uid, err := uuid.Parse(id)
	if err != nil {
		a.writeResponseError(w, fmt.Errorf("invalid parameter passed: %s", err), http.StatusBadRequest)
		return
	}

	u, err := a.store.GetUserByID(uid)
	if err != nil {
		a.writeResponseError(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(u)
}

func (a *API) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminUser(r.Context()) {
		a.writeResponseError(w, ErrPermissionsDenied, http.StatusForbidden)
		return
	}

	id := mux.Vars(r)["id"]
	uid, err := uuid.Parse(id)
	if err != nil {
		a.writeResponseError(w, fmt.Errorf("invalid parameter passed: %s", err), http.StatusBadRequest)
		return
	}

	var u database.User
	err = json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		a.writeResponseError(w, fmt.Errorf("wrong JSON: %s", err), http.StatusBadRequest)
		return
	}

	if u.Email != "" {
		if err := validate.Var(u.Email, "email"); err != nil {
			a.writeResponseError(w, fmt.Errorf("invalid data passed: %s", err), http.StatusBadRequest)
		}
	}

	u.ID = uid

	err = a.store.UpdateUser(&u)
	if err != nil {
		if errors.Is(err, database.ErrUserNotExist) || errors.Is(err, database.ErrNameAlreadyExist) {
			a.writeResponseError(w, err, http.StatusBadRequest)
			return
		}
		a.internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminUser(r.Context()) {
		a.writeResponseError(w, ErrPermissionsDenied, http.StatusForbidden)
		return
	}

	id := mux.Vars(r)["id"]
	uid, err := uuid.Parse(id)
	if err != nil {
		a.writeResponseError(w, fmt.Errorf("invalid parameter passed: %s", err), http.StatusBadRequest)
		return
	}

	err = a.store.DeleteUser(uid)
	if err != nil {
		a.writeResponseError(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
