package store

import (
	"database/sql"
	"fmt"

	"github.com/heroku/stocksignals/model"
)

// RegisterUser registers the given user to the database, if it doesn't exist with that email
func RegisterUser(user model.User) error {
	if db == nil {
		return fmt.Errorf("no connection is created to the database")
	}

	if user.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if len(user.Password) < 5 {
		return fmt.Errorf("password length must be at least 5 characters")
	}

	var result model.User
	err := db.Get(&result, fmt.Sprintf("SELECT * FROM users WHERE email='%s'", user.Email))
	fmt.Println("Email : ", user.Email)
	if err == sql.ErrNoRows {
		_, errRegister := db.NamedExec("INSERT INTO users (email, password) VALUES (:email, :password)", &user)
		if errRegister != nil {
			return fmt.Errorf("error registering user with email %s: %q", user.Email, err)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading user with email %s: %q", user.Email, err)
	}

	return fmt.Errorf("user already exists with email %s", user.Email)
}

// GetUser gets the user with the given email
func GetUser(email string) (*model.User, error) {
	if db == nil {
		return nil, fmt.Errorf("no connection is created to the database")
	}

	if email == "" {
		return nil, fmt.Errorf("empty email cannot be queried")
	}

	var result model.User
	err := db.Get(&result, fmt.Sprintf("SELECT * FROM users WHERE email='%s'", email))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user with email %s does not exist.", email)
	}

	if err != nil {
		return nil, fmt.Errorf("error reading user with email %s: %q", email, err)
	}

	return &result, nil
}

// GetUsers gets all the users
func GetUsers() ([]model.User, error) {
	if db == nil {
		return nil, fmt.Errorf("no connection is created to the database")
	}

	var results []model.User
	err := db.Select(&results, "SELECT * FROM users")
	if err != nil {
		return nil, fmt.Errorf("error reading users: %q", err)
	}

	return results, nil
}
