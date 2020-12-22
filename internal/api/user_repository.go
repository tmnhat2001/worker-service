package api

import "errors"

// UserRepository defines an interface to retrieve stored users
type UserRepository interface {
	FindByUsername(username string) (*User, error)
}

// MemoryUserRepository stores users in memory and implements UserRepository
type MemoryUserRepository struct {
	Users map[string]*User
}

// FindByUsername returns a user given its username. If no such user exists, returns an error
func (repository *MemoryUserRepository) FindByUsername(username string) (*User, error) {
	user, ok := repository.Users[username]
	if !ok {
		return nil, errors.New("Cannot find user that matches username")
	}

	return user, nil
}
