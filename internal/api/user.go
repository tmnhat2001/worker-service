package api

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user that has access to the API
type User struct {
	Username     string
	PasswordHash []byte
}

// createUsers creates dummy users for manual testing
func createUsers() (map[string]*User, error) {
	users := make(map[string]*User)
	usernames := []string{"user1", "user2"}
	for _, username := range usernames {
		password := fmt.Sprintf("thisispasswordfor%s", username)
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCostFactor)
		if err != nil {
			return nil, err
		}

		user := User{Username: username, PasswordHash: passwordHash}
		users[username] = &user
	}

	return users, nil
}
