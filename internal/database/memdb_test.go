package database

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUser_CheckPassword(t *testing.T) {
	password := "qwerty"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	u := User{
		Password: string(hash),
	}

	assert.True(t, u.CheckPassword(password), "Checking password should be correct")
}

func TestUser_UpdateFields_ChangePassword(t *testing.T) {
	id := uuid.New()
	newPassword := "12345"
	oldHashedPassword, _ := bcrypt.GenerateFromPassword([]byte("qwery"), bcrypt.DefaultCost)

	oldUser := &User{
		ID:       id,
		Password: string(oldHashedPassword),
	}

	newUser := &User{
		ID:       id,
		Password: newPassword,
	}

	_ = newUser.UpdateFields(oldUser)

	assert.True(t, newUser.CheckPassword(newPassword))
}

func TestUser_UpdateFields_CopyFields(t *testing.T) {
	id := uuid.New()
	username := "1"
	email := "e@mail.ru"
	oldHashedPassword, _ := bcrypt.GenerateFromPassword([]byte("qwery"), bcrypt.DefaultCost)

	oldUser := &User{
		ID:       id,
		Username: username,
		Password: string(oldHashedPassword),
		Email:    email,
		Admin:    true,
	}

	newUser := &User{
		ID: id,
	}

	wantUser := &User{
		ID:       id,
		Username: username,
		Password: string(oldHashedPassword),
		Email:    email,
		Admin:    false,
	}

	_ = newUser.UpdateFields(oldUser)

	assert.Equal(t, wantUser, newUser)
}

func TestDB_NewUser_ErrNameAlreadyExist(t *testing.T) {
	db := New()

	username := "1"

	err := db.NewUser(&User{
		Username: username,
	})
	assert.Nil(t, err, "Database shouldn't raise an error on first user creation")

	err = db.NewUser(&User{
		Username: username,
	})

	assert.ErrorIs(t, err, ErrNameAlreadyExist)
}

func TestDB_GetAllUsers(t *testing.T) {
	db := New()

	user := &User{
		Username: "1",
	}
	err := db.NewUser(user)
	assert.Nil(t, err)

	users := db.GetAllUsers()
	assert.Equal(t, user, users[0])
	assert.Equal(t, 1, len(users))
}

func TestDB_GetUserByName_ErrUserNotExist(t *testing.T) {
	db := New()

	err := db.NewUser(&User{
		Username: "1",
	})
	assert.Nil(t, err)

	_, err = db.GetUserByName("2")
	assert.ErrorIs(t, err, ErrUserNotExist)
}

func TestDB_GetUserByName_GoodWay(t *testing.T) {
	db := New()

	username := "1"
	user := &User{
		Username: username,
	}

	err := db.NewUser(user)
	assert.Nil(t, err)

	gotUser, _ := db.GetUserByName(username)
	assert.Equal(t, user, gotUser)
}

func TestDB_GetUserByID_ErrUserNotExist(t *testing.T) {
	db := New()

	err := db.NewUser(&User{
		Username: "1",
	})
	assert.Nil(t, err)

	fakeKey := uuid.New()

	_, err = db.GetUserByID(fakeKey)
	assert.ErrorIs(t, err, ErrUserNotExist)
}

func TestDB_GetUserByID_GoodWay(t *testing.T) {
	db := New()

	user := &User{
		Username: "1",
	}

	err := db.NewUser(user)
	assert.Nil(t, err)

	gotUser, _ := db.GetUserByID(user.ID)
	assert.Equal(t, user, gotUser)
}

func TestDB_UpdateUser_ErrUserNotExist(t *testing.T) {
	db := New()

	updateUser := &User{
		ID:       uuid.New(),
		Username: "new",
	}

	err := db.UpdateUser(updateUser)
	assert.ErrorIs(t, err, ErrUserNotExist)
}

func TestDB_UpdateUser_ErrNameAlreadyExist(t *testing.T) {
	db := New()

	firstUserName := "exist"

	err := db.NewUser(&User{
		Username: firstUserName,
	})
	assert.Nil(t, err)

	secondUser := &User{
		Username: "second",
	}
	err = db.NewUser(secondUser)
	assert.Nil(t, err)

	updateUser := &User{
		ID:       secondUser.ID,
		Username: firstUserName,
	}

	err = db.UpdateUser(updateUser)
	assert.ErrorIs(t, err, ErrNameAlreadyExist)
}

func TestDB_UpdateUser_GoodWay(t *testing.T) {
	db := New()

	wantUsername := "want"
	wantEmail := "want@e.mail"

	oldUser := &User{
		Username: wantUsername,
		Email:    "e@mai.l",
	}

	err := db.NewUser(oldUser)
	assert.Nil(t, err)

	wantID := oldUser.ID
	wantPassword := oldUser.Password

	newUser := &User{
		ID:    wantID,
		Email: wantEmail,
	}

	wantUser := &User{
		ID:       wantID,
		Username: wantUsername,
		Email:    wantEmail,
		Password: wantPassword,
	}

	err = db.UpdateUser(newUser)
	assert.Nil(t, err)

	gotUser, err := db.GetUserByID(wantID)
	assert.Nil(t, err)
	assert.Equal(t, wantUser, gotUser)
}

func TestDB_DeleteUser_ErrUserNotExist(t *testing.T) {
	db := New()

	err := db.NewUser(&User{
		ID:       uuid.New(),
		Username: "new",
	})
	assert.Nil(t, err)

	fakeKey := uuid.New()

	err = db.DeleteUser(fakeKey)
	assert.ErrorIs(t, err, ErrUserNotExist)
}

func TestDB_DeleteUser_GoodWay(t *testing.T) {
	db := New()

	user := &User{
		Username: "1",
	}

	err := db.NewUser(user)
	assert.Nil(t, err)

	err = db.DeleteUser(user.ID)
	assert.Nil(t, err)

	users := db.GetAllUsers()
	assert.Equal(t, 0, len(users))
}
