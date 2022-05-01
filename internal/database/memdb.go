package database

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

var ErrNameAlreadyExist error = errors.New("this name already exists")
var ErrUserNotExist error = errors.New("user does not exist")

type User struct {
	ID       uuid.UUID
	Email    string `validate:"email"`
	Username string `validate:"min=1"`
	Password string `validate:"min=1"`
	Admin    bool
}

type DB struct {
	mu     sync.Mutex
	store  map[uuid.UUID]User
	unames map[string]uuid.UUID
}

func New() *DB {
	return &DB{
		mu:     sync.Mutex{},
		unames: map[string]uuid.UUID{},
		store:  map[uuid.UUID]User{},
	}
}

//NewUser creates a user, returns id.
func (db *DB) NewUser(u User) (uuid.UUID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	uid := uuid.New() //uuid должен сгенерироваться уникальный, но существует малая вероятность коллизии, поэтому добавлю проверку
	if _, ok := db.store[uid]; ok {
		return uuid.Nil, errors.New("id is not unique")
	}

	if _, ok := db.unames[u.Username]; ok {
		return uuid.Nil, ErrNameAlreadyExist
	}

	u.ID = uid
	db.unames[u.Username] = uid
	db.store[uid] = u

	return uid, nil
}

//GetAllUsers returns a list of all users.
func (db *DB) GetAllUsers() []User {
	db.mu.Lock()
	defer db.mu.Unlock()

	var users []User
	for _, u := range db.store {
		users = append(users, u)
	}

	return users
}

//GetUserByName finds a user by name.
func (db *DB) GetUserByName(uname string) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	id, ok := db.unames[uname]
	if !ok {
		return User{}, ErrUserNotExist
	}

	return db.store[id], nil
}

//GetUserByID finds a user by ID.
func (db *DB) GetUserByID(uid uuid.UUID) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	u, ok := db.store[uid]
	if !ok {
		return User{}, ErrUserNotExist
	}

	return u, nil
}

//UpdateUser updates user data. The username must be unique.
func (db *DB) UpdateUser(u User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	user, ok := db.store[u.ID]
	if !ok {
		return ErrUserNotExist
	}

	if u.Username != user.Username {
		if _, ok := db.unames[u.Username]; ok {
			return ErrNameAlreadyExist
		}

		delete(db.unames, user.Username)
		db.unames[u.Username] = u.ID
	}

	db.store[u.ID] = u
	return nil
}

//DeleteUser deletes a user by ID.
func (db *DB) DeleteUser(uid uuid.UUID) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	user, ok := db.store[uid]
	if !ok {
		return ErrUserNotExist
	}

	delete(db.unames, user.Username)
	delete(db.store, uid)

	return nil
}
