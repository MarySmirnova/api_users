package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MarySmirnova/api_users/internal/config"
	"github.com/MarySmirnova/api_users/internal/database"
)

const (
	adminUname string = "Admin"
	adminPass  string = "Admin"
)

func TestAPI_newUserHandler(t *testing.T) {

	db := database.New()
	_, err := db.NewUser(database.User{
		Username: adminUname,
		Password: adminPass,
		Admin:    true,
	})

	if err != nil {
		t.Errorf("error: %s", err)
	}

	api := New(config.API{
		Listen:       ":8080",
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}, db)

	newUser := database.User{}
	b, _ := json.Marshal(newUser)

	req, _ := http.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(b))
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("wrong code: got %d, want %d", resp.Code, http.StatusBadRequest)
	}
}

func execRequest(req *http.Request, s *http.Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()

	s.Handler.ServeHTTP(rr, req)
	return rr
}
