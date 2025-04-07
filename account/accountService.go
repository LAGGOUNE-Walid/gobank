package account

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Entity struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Number    int       `json:"number"`
	Balance   int       `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type CreateResponse struct {
	ID int `json:"id"`
}

type Response struct {
	Data  []Entity `json:"data"`
	Page  int      `json:"page"`
	Limit int      `json:"limit"`
	Total int      `json:"total"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID    int    `json:"id"`
	Token string `json:"token"`
}

type TransferRequest struct {
	To     int `json:"to"`
	Ammout int `json:"ammount"`
}

func New(username string, password string, firstname string, lastname string) (*Entity, error) {
	if len(username) == 0 || len(password) == 0 || len(firstname) == 0 || len(lastname) == 0 {
		return nil, fmt.Errorf("firstname or lastname validation failed")
	}
	passwordHashed, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash your password :(")
	}
	return &Entity{
		Username:  username,
		Firstname: firstname,
		Lastname:  lastname,
		Password:  passwordHashed,
		Number:    rand.Intn(10000),
		CreatedAt: time.Now().UTC(),
	}, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
