package api

import (
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// BcryptCostFactor is the cost factor for hashing passwords with bcrypt
const BcryptCostFactor = 12

// AuthenticationService is a service to authenticate users
type AuthenticationService struct {
	UserRepository UserRepository
}

// Authenticate checks if the credentials of the given Request is valid
func (service *AuthenticationService) Authenticate(req *http.Request) (*User, error) {
	username, password, ok := req.BasicAuth()
	if !ok {
		return nil, errors.New("Unable to parse Basic Auth credentials from request")
	}

	user, err := service.UserRepository.FindByUsername(username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}
