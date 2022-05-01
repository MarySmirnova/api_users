package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MarySmirnova/api_users/internal/database"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var validate = validator.New()

func (a *API) newUserHandler(w http.ResponseWriter, r *http.Request) {
	admin, ok := r.Context().Value(isAdmin).(bool)
	if !ok {
		callError(w, ErrWrongContext, http.StatusInternalServerError)
		return
	}
	if !admin {
		callError(w, ErrPermissionsDenied, http.StatusForbidden)
		return
	}

	var u database.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		callError(w, fmt.Errorf("wrong JSON: %s", err), http.StatusBadRequest)
		return
	}

	err = validate.Struct(u)
	if err != nil {
		callError(w, fmt.Errorf("invalid data passed: %s", err), http.StatusBadRequest)
		return
	}

	id, err := a.store.NewUser(u)
	if err != nil {
		if err == database.ErrNameAlreadyExist {
			callError(w, err, http.StatusBadRequest)
			return
		}
		callError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(id)
}

func (a *API) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	u := a.store.GetAllUsers()

	if len(u) == 0 {
		callError(w, ErrEmptyDB, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func (a *API) getUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	uid, err := uuid.Parse(id)
	if err != nil {
		callError(w, fmt.Errorf("invalid parameter passed: %s", err), http.StatusBadRequest)
		return
	}

	u, err := a.store.GetUserByID(uid)
	if err != nil {
		callError(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func (a *API) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	admin, ok := r.Context().Value(isAdmin).(bool)
	if !ok {
		callError(w, ErrWrongContext, http.StatusInternalServerError)
		return
	}
	if !admin {
		callError(w, ErrPermissionsDenied, http.StatusForbidden)
		return
	}

	id := mux.Vars(r)["id"]
	uid, err := uuid.Parse(id)
	if err != nil {
		callError(w, fmt.Errorf("invalid parameter passed: %s", err), http.StatusBadRequest)
		return
	}

	u, err := a.store.GetUserByID(uid)
	if err != nil {
		callError(w, err, http.StatusBadRequest)
		return
	}

	var newData database.User
	err = json.NewDecoder(r.Body).Decode(&newData)
	if err != nil {
		callError(w, fmt.Errorf("wrong JSON: %s", err), http.StatusBadRequest)
		return
	}

	if newData.Username != "" && newData.Username != u.Username {
		u.Username = newData.Username
	}
	if newData.Password != "" && newData.Password != u.Password {
		u.Password = newData.Password
	}
	if newData.Email != "" && newData.Email != u.Email {
		u.Email = newData.Email
	}
	u.Admin = newData.Admin

	err = validate.Struct(u)
	if err != nil {
		callError(w, fmt.Errorf("invalid data passed: %s", err), http.StatusBadRequest)
		return
	}

	err = a.store.UpdateUser(u)
	if err != nil {
		callError(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	admin, ok := r.Context().Value(isAdmin).(bool)
	if !ok {
		callError(w, ErrWrongContext, http.StatusInternalServerError)
		return
	}
	if !admin {
		callError(w, ErrPermissionsDenied, http.StatusForbidden)
		return
	}

	id := mux.Vars(r)["id"]
	uid, err := uuid.Parse(id)
	if err != nil {
		callError(w, fmt.Errorf("invalid parameter passed: %s", err), http.StatusBadRequest)
		return
	}

	err = a.store.DeleteUser(uid)
	if err != nil {
		callError(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
