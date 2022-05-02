package database

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

//CheckPassword compares a hashed password with string password.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))

	return err == nil
}

//CreatePasswordHash creates a hashed password from a string.
func (u *User) CreatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

//UpdateFields updates empty fields in the struct with data from the passed struct.
//When changing the password, hashes it.
func (u *User) UpdateFields(oldUser *User) error {
	if u.Username == "" {
		u.Username = oldUser.Username
	}
	if u.Email == "" {
		u.Email = oldUser.Email
	}
	if u.Password == "" {
		u.Password = oldUser.Password
		return nil
	}

	newHashedPass, err := u.CreatePasswordHash(u.Password)
	if err != nil {
		return err
	}
	u.Password = newHashedPass

	return nil
}

type DB struct {
	mu            sync.RWMutex
	unamesUniqKey map[string]uuid.UUID
	store         map[uuid.UUID]*User
}

func New() *DB {
	return &DB{
		mu:            sync.RWMutex{},
		unamesUniqKey: make(map[string]uuid.UUID),
		store:         make(map[uuid.UUID]*User),
	}
}

//NewUser creates a user, returns id.
func (db *DB) NewUser(u *User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	u.ID = uuid.New() //uuid должен сгенерироваться уникальный, но существует малая вероятность коллизии, поэтому добавлю проверку
	if _, ok := db.store[u.ID]; ok {
		return errors.New("id is not unique")
	}

	if _, ok := db.unamesUniqKey[u.Username]; ok {
		return ErrNameAlreadyExist
	}

	hashedPass, err := u.CreatePasswordHash(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPass

	db.unamesUniqKey[u.Username] = u.ID
	db.store[u.ID] = u

	return nil
}

//GetAllUsers returns a list of all users.
func (db *DB) GetAllUsers() []*User {
	db.mu.RLock()
	defer db.mu.RUnlock()

	users := make([]*User, 0, len(db.store))
	for _, u := range db.store {
		users = append(users, u)
	}

	return users
}

//GetUserByName finds a user by name.
func (db *DB) GetUserByName(uname string) (*User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	id, ok := db.unamesUniqKey[uname]
	if !ok {
		return nil, ErrUserNotExist
	}

	return db.store[id], nil
}

//GetUserByID finds a user by ID.
func (db *DB) GetUserByID(uid uuid.UUID) (*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	u, ok := db.store[uid]
	if !ok {
		return nil, ErrUserNotExist
	}

	return u, nil
}

//UpdateUser updates user data. The username must be unique.
func (db *DB) UpdateUser(u *User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	user, ok := db.store[u.ID]
	if !ok {
		return ErrUserNotExist
	}

	if err := u.UpdateFields(user); err != nil {
		return err
	}

	if u.Username != user.Username {
		if _, ok := db.unamesUniqKey[u.Username]; ok {
			return ErrNameAlreadyExist
		}

		delete(db.unamesUniqKey, user.Username)
		db.unamesUniqKey[u.Username] = u.ID
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

	delete(db.unamesUniqKey, user.Username)
	delete(db.store, uid)

	return nil
}
