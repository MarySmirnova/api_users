package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MarySmirnova/api_users/internal/config"
	"github.com/MarySmirnova/api_users/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	adminUname string = "IsAdmin"
	adminPass  string = "IsAdmin"

	notAdminUname string = "IsNotAdmin"
	notAdminPass  string = "IsNotAdmin"
)

func testBootstrap(t *testing.T) (*API, uuid.UUID) {
	db := database.New()
	err := db.NewUser(&database.User{
		Username: adminUname,
		Password: adminPass,
		Admin:    true,
	})
	assert.Nil(t, err, "Database shouldn't raise an error on first user creation")

	user := &database.User{
		Username: notAdminUname,
		Password: notAdminPass,
		Admin:    false,
	}
	err = db.NewUser(user)
	assert.Nil(t, err)

	api := New(config.API{
		Listen:       ":8080",
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}, db)

	return api, user.ID
}

func toJSON(v interface{}) io.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func execRequest(req *http.Request, s *http.Server) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	s.Handler.ServeHTTP(res, req)

	return res
}

func TestAPI_NewUserHandler_InvalidFileds(t *testing.T) {
	api, _ := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodPost, "/user", toJSON(database.User{}))
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestAPI_NewUserHandler_WrongJSON(t *testing.T) {
	api, _ := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodPost, "/user", toJSON("User"))
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestAPI_NewUserHandler_ErrNameAlreadyExist(t *testing.T) {
	api, _ := testBootstrap(t)

	user := database.User{
		Username: adminUname,
		Password: adminPass,
		Email:    "e@mail.ru",
	}

	req, _ := http.NewRequest(http.MethodPost, "/user", toJSON(user))
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, string(resp.Body.Bytes()), database.ErrNameAlreadyExist.Error())
}

func TestAPI_NewUserHandler_PermissionsDenied(t *testing.T) {
	api, _ := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodPost, "/user", toJSON(database.User{}))
	req.SetBasicAuth(notAdminUname, notAdminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestAPI_NewUserHandler_GoodWay(t *testing.T) {
	api, _ := testBootstrap(t)

	user := database.User{
		Email:    "this@is.the",
		Username: "good",
		Password: "way",
	}

	req, _ := http.NewRequest(http.MethodPost, "/user", toJSON(user))
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusOK, resp.Code)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	var id string
	err = json.Unmarshal(body, &id)
	assert.Nil(t, err)

	_, err = uuid.Parse(id)
	assert.Nil(t, err)
}

func TestAPI_GetUsersHandler(t *testing.T) {
	api, _ := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodGet, "/user", nil)
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusOK, resp.Code)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	var data []database.User
	err = json.Unmarshal(body, &data)
	assert.Nil(t, err)

	wantLen := 2
	assert.Equal(t, wantLen, len(data))
}

func TestAPI_GetUserByIDHandler_InvalidID(t *testing.T) {
	api, _ := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodGet, "/user/{1}", nil)
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestAPI_GetUserByIDHandler_ErrUserNotExist(t *testing.T) {
	api, _ := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/user/{%s}", uuid.New()), nil)
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, string(resp.Body.Bytes()), database.ErrUserNotExist.Error())
}

func TestAPI_GetUserByIDHandler_GoodWay(t *testing.T) {
	api, id := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/user/{%s}", id), nil)
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusOK, resp.Code)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	var data database.User
	err = json.Unmarshal(body, &data)
	assert.Nil(t, err)

	assert.Equal(t, notAdminUname, data.Username)
}

func TestAPI_UpdateUserHandler_PermissionsDenied(t *testing.T) {
	api, id := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/user/{%s}", id), toJSON(database.User{}))
	req.SetBasicAuth(notAdminUname, notAdminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestAPI_UpdateUserHandler_InvalidID(t *testing.T) {
	api, _ := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodPatch, "/user/{1}", toJSON(database.User{}))
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestAPI_UpdateUserHandler_WrongJSON(t *testing.T) {
	api, id := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/user/{%s}", id), toJSON("User"))
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestAPI_UpdateUserHandler_InvalidEmail(t *testing.T) {
	api, id := testBootstrap(t)

	user := database.User{
		Email: "qwerty",
	}

	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/user/{%s}", id), toJSON(user))
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestAPI_UpdateUserHandler_ErrUserNotExist(t *testing.T) {
	api, _ := testBootstrap(t)

	user := database.User{
		Email: "qwerty",
	}

	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/user/{%s}", uuid.New()), toJSON(user))
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, string(resp.Body.Bytes()), database.ErrUserNotExist.Error())
}

func TestAPI_UpdateUserHandler_ErrNameAlreadyExist(t *testing.T) {
	api, id := testBootstrap(t)

	user := database.User{
		Username: adminUname,
		Email:    "e@mail.ru",
	}

	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/user/{%s}", id), toJSON(user))
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, string(resp.Body.Bytes()), database.ErrNameAlreadyExist.Error())
}

func TestAPI_UpdateUserHandler_GoodWay(t *testing.T) {
	api, id := testBootstrap(t)

	user := database.User{
		Email:    "e@mail.ru",
		Password: "12345",
	}

	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/user/{%s}", id), toJSON(user))
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusNoContent, resp.Code)
}

func TestAPI_DeleteUserHandler_PermissionsDenied(t *testing.T) {
	api, id := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/user/{%s}", id), nil)
	req.SetBasicAuth(notAdminUname, notAdminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestAPI_DeleteUserHandler_InvalidID(t *testing.T) {
	api, _ := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodDelete, "/user/{1}", nil)
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestAPI_DeleteUserHandler_ErrUserNotExist(t *testing.T) {
	api, _ := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/user/{%s}", uuid.New()), nil)
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, string(resp.Body.Bytes()), database.ErrUserNotExist.Error())
}

func TestAPI_DeleteUserHandler_GoodWay(t *testing.T) {
	api, id := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/user/{%s}", id), nil)
	req.SetBasicAuth(adminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusNoContent, resp.Code)
}

func TestAPI_AuthMiddleware_NoAuth(t *testing.T) {
	api, _ := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodGet, "/user", nil)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestAPI_AuthMiddleware_WrongPass(t *testing.T) {
	api, _ := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodGet, "/user", nil)
	req.SetBasicAuth(notAdminUname, adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestAPI_AuthMiddleware_WrongUsername(t *testing.T) {
	api, _ := testBootstrap(t)

	req, _ := http.NewRequest(http.MethodGet, "/user", nil)
	req.SetBasicAuth("tolik", adminPass)

	resp := execRequest(req, api.httpServer)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}
