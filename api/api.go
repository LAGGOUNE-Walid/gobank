package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/LAGGOUNE-Walid/gobank/account"
	"github.com/LAGGOUNE-Walid/gobank/storage"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("fake-secret-key")

type contextKey string

const userIDKey = contextKey("userID")

type Server struct {
	addr  string
	store storage.AccountStore
}

type ApiError struct {
	Error string
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func (s *Server) Run() {
	mux := s.Routes()
	http.ListenAndServe(s.addr, mux)
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/account", makeHttpHandlerFunc(s.handleAccount))
	mux.HandleFunc("/account/{id}", makeHttpHandlerFunc(WithAuth(s.handleSingleAccount)))
	mux.HandleFunc("/transfer", makeHttpHandlerFunc(WithAuth(s.handleTransferAccount)))
	mux.HandleFunc("/login", makeHttpHandlerFunc(s.handleLogin))
	return mux
}

func NewServer(addr string, store storage.AccountStore) *Server {
	return &Server{
		addr:  addr,
		store: store,
	}
}

func (s *Server) handleAccount(writer http.ResponseWriter, request *http.Request) error {
	if request.Method == "GET" {
		return s.handleGetAccounts(writer, request)
	}
	if request.Method == "POST" {
		return s.handleCreateAccount(writer, request)
	}

	return fmt.Errorf("HTTP method not allowed %s", request.Method)
}

func (s *Server) handleGetAccounts(writer http.ResponseWriter, request *http.Request) error {
	pageQuery := request.URL.Query().Get("page")
	page, err := stringToInt(pageQuery)
	if err != nil {
		return err
	}
	accounts, err := s.store.All(20, page)
	if err != nil {
		return err
	}
	return writeJson(writer, http.StatusOK, accounts)
}

func (s *Server) handleSingleAccount(writer http.ResponseWriter, request *http.Request) error {
	if request.Method == "DELETE" {
		return s.handleDeleteAccount(writer, request)
	}
	if request.Method == "GET" {
		return s.handleGetAccount(writer, request)
	}

	return fmt.Errorf("HTTP method not allowed %s", request.Method)
}

func (s *Server) handleGetAccount(writer http.ResponseWriter, request *http.Request) error {
	id, err := stringToInt(request.PathValue("id"))
	if err != nil {
		return err
	}
	if id != request.Context().Value(userIDKey) {
		return writeJson(writer, http.StatusUnauthorized, nil)
	}
	account, err := s.store.Find(id)
	if err != nil {
		return err
	}
	return writeJson(writer, http.StatusOK, account)
}

func (s *Server) handleCreateAccount(writer http.ResponseWriter, request *http.Request) error {
	createAccountRequest := new(account.CreateRequest)
	if err := json.NewDecoder(request.Body).Decode(createAccountRequest); err != nil {
		return err
	}
	accountEntity, err := account.New(createAccountRequest.Username, createAccountRequest.Password, createAccountRequest.Firstname, createAccountRequest.Lastname)
	if err != nil {
		return err
	}
	id, err := s.store.Create(accountEntity)
	if err != nil {
		return err
	}
	return writeJson(writer, http.StatusCreated, account.CreateResponse{ID: id})
}

func (s *Server) handleDeleteAccount(writer http.ResponseWriter, request *http.Request) error {
	id, err := stringToInt(request.PathValue("id"))
	if err != nil {
		return err
	}
	if id != request.Context().Value(userIDKey) {
		return writeJson(writer, http.StatusUnauthorized, nil)
	}

	s.store.Delete(id)
	return writeJson(writer, http.StatusCreated, nil)
}

func (s *Server) handleTransferAccount(writer http.ResponseWriter, request *http.Request) error {
	transferRequest := new(account.TransferRequest)
	if request.Method != "POST" {
		return fmt.Errorf("not supported %s method , only POST method supported", request.Method)
	}
	if err := json.NewDecoder(request.Body).Decode(&transferRequest); err != nil {
		return err
	}
	if transferRequest.Ammout <= 0 {
		return fmt.Errorf("can't transfer ammount <= 0")
	}
	err := s.store.Transfer(transferRequest.To, request.Context().Value(userIDKey).(int), transferRequest.Ammout)
	if err != nil {
		return err
	}
	return writeJson(writer, http.StatusCreated, "")
}

func (s *Server) handleLogin(writer http.ResponseWriter, request *http.Request) error {
	var loginRequest account.LoginRequest
	if request.Method != "POST" {
		return fmt.Errorf("not supported %s method , only POST method supported", request.Method)
	}
	if err := json.NewDecoder(request.Body).Decode(&loginRequest); err != nil {
		return err
	}
	user, err := s.store.Login(loginRequest.Username, loginRequest.Password)
	if err != nil {
		return writeJson(writer, http.StatusBadRequest, "invalid credentials")
	}
	tokenString, err := createToken(user.ID, user.Username)
	if err != nil {
		return err
	}
	return writeJson(writer, http.StatusOK, account.LoginResponse{Token: tokenString, ID: user.ID})
}

func makeHttpHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		if err := f(writer, request); err != nil {
			writeJson(writer, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func WithAuth(next apiFunc) apiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			return fmt.Errorf("missing authorization header")
		}

		tokenString = tokenString[len("Bearer "):]
		userID, err := verifyToken(tokenString)
		if err != nil {
			return err
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		r = r.WithContext(ctx)
		return next(w, r)
	}
}

func writeJson(writer http.ResponseWriter, status int, v any) error {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	return json.NewEncoder(writer).Encode(v)
}

func stringToInt(v string) (int, error) {
	if newV, err := strconv.Atoi(v); err == nil && newV > 0 {
		return newV, nil
	}
	return 0, fmt.Errorf("failed to convert %s to integer", v)
}

func createToken(id int, username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"id":       id,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func verifyToken(tokenString string) (userID int, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("cannot parse claims")
	}

	id, ok := claims["id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id not found in token")
	}
	userID = int(id)
	return userID, nil
}
