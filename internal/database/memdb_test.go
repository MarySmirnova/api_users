package database

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestDB_User(t *testing.T) {
	db := New()
	u := User{
		Email:    "mail.ru",
		Username: "Admin",
		Password: "qwerty",
		Admin:    true,
	}

	//testing NewUser
	id, _ := db.NewUser(u)
	u.ID = id

	_, err := db.NewUser(u)
	if err != ErrNameAlreadyExist {
		t.Errorf("incorrect error: got %v, want %v", err, ErrNameAlreadyExist)
	}

	//testing GetUserByID
	user, _ := db.GetUserByID(id)
	if !reflect.DeepEqual(user, u) {
		t.Errorf("created user not found: got %v, want %v", user, u)
	}

	fakeID := uuid.New()
	_, err = db.GetUserByID(fakeID)
	if err != ErrUserNotExist {
		t.Errorf("incorrect error: got %v, want %v", err, ErrUserNotExist)
	}

	//testing UpdateUser
	id2, _ := db.NewUser(User{
		Username: "NotAdmin",
		Password: "qwerty",
	})

	err = db.UpdateUser(User{
		ID:       id2,
		Username: "Admin",
	})
	if err != ErrNameAlreadyExist {
		t.Errorf("incorrect error: got %v, want %v", err, ErrNameAlreadyExist)
	}

	err = db.UpdateUser(User{
		ID:       fakeID,
		Username: "Kolya",
	})
	if err != ErrUserNotExist {
		t.Errorf("incorrect error: got %v, want %v", err, ErrUserNotExist)
	}

	wantPass := "12345"
	db.UpdateUser(User{
		ID:       id2,
		Username: "NotAdmin",
		Password: wantPass,
	})
	u2, _ := db.GetUserByID(id2)
	if !reflect.DeepEqual(u2.Password, wantPass) {
		t.Errorf("failed to update: got %s, want %s", u2.Password, wantPass)
	}

	wantName := "Kolya"
	db.UpdateUser(User{
		ID:       id2,
		Username: wantName,
		Password: wantPass,
	})
	u2, _ = db.GetUserByID(id2)
	if !reflect.DeepEqual(u2.Username, wantName) {
		t.Errorf("failed to update: got %s, want %s", u2.Username, wantName)
	}

	//testing DeleteUser and GetAllUsers
	err = db.DeleteUser(fakeID)
	if err != ErrUserNotExist {
		t.Errorf("incorrect error: got %v, want %v", err, ErrUserNotExist)
	}

	db.DeleteUser(id2)
	users := db.GetAllUsers()
	if len(users) != 1 {
		t.Errorf("failed to delete")
	}

	//testing GetUserByName
	_, err = db.GetUserByName("Kolya")
	if err != ErrUserNotExist {
		t.Errorf("incorrect error: got %v, want %v", err, ErrUserNotExist)
	}

	gotUser, _ := db.GetUserByName("Admin")
	if !reflect.DeepEqual(gotUser, u) {
		t.Errorf("incorrect data: got %v, want %v", gotUser, u)
	}
}
