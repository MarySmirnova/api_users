package api

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var (
	ErrPermissionsDenied error = errors.New("insufficient permissions")
	ErrWrongContext      error = errors.New("unable to read context")
	ErrUnauthorized      error = errors.New("unauthorized")
	ErrEmptyDB           error = errors.New("database is empty")
)

func callError(w http.ResponseWriter, err error, code int) {
	log.WithError(err).Error("api error")
	http.Error(w, err.Error(), code)
}
