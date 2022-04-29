package database

import "sync"

type User struct {
	ID       int
	Email    string
	Username string
	Password string
	Admin    bool
}

type DB struct {
	mu    sync.Mutex
	id    int
	store map[int]User
}

func New() *DB {
	return &DB{
		mu:    sync.Mutex{},
		id:    1,
		store: make(map[int]User),
	}
}

//NewUser creates a user, returns id.
func (db *DB) NewUser(u User) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	return 0, nil
}
