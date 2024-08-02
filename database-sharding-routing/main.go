package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	mux := http.NewServeMux()
	db1 := createDb("localhost", 5430, "postgres", "postgres", "postgres", 10)
	setupDatabaseTables(db1)
	db2 := createDb("localhost", 5431, "postgres", "postgres", "postgres", 10)
	setupDatabaseTables(db2)

	dbManager := NewDbManager(db1, db2)
	userService := NewUserService(dbManager)

	mux.Handle("POST /auth", http.HandlerFunc(login(userService)))
	mux.Handle("POST /users", http.HandlerFunc(createUser(userService)))
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err) // log.Fatal will log the error and exit the program
	}
}

type DbManager struct {
	db1 *sql.DB
	db2 *sql.DB
}

func NewDbManager(db1, db2 *sql.DB) *DbManager {
	dbm := &DbManager{db1: db1, db2: db2}
	return dbm
}

func (dm *DbManager) GetDb(username string) *sql.DB {
	firstChar := strings.ToLower(username)[0]
	if firstChar >= 'a' && firstChar <= 'm' {
		log.Printf("Routing %s to db1", username)
		return dm.db1
	}
	log.Printf("Routing %s to db2", username)
	return dm.db2
}

type UserService struct {
	db *DbManager
}

func NewUserService(db *DbManager) *UserService {
	return &UserService{db: db}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func login(s *UserService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		db := s.db.GetDb(req.Username)
		var hashedPass string
		err = db.QueryRow("SELECT password FROM users WHERE username = $1", req.Username).Scan(&hashedPass)
		if err != nil {
			log.Printf("error getting user: %v", err)
			http.Error(w, "invalid username or password", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(req.Password))
		if err != nil {
			log.Printf("invalid password for user %s", req.Username)
			http.Error(w, "invalid username or password", http.StatusUnauthorized)
			return
		}

		log.Printf("user %s logged in", req.Username)
		w.WriteHeader(http.StatusOK)
	}
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func createUser(s *UserService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		// Create the user in the database
		db := s.db.GetDb(req.Username)
		// make password storable by hashing it
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("error hashing password: %v", err)
			http.Error(w, fmt.Sprintf("error hashing password: %v", err), http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)", req.Username, req.Email, string(hashedPass))
		if err != nil {
			log.Printf("error inserting user: %v", err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		log.Printf("user %s created", req.Username)
		w.WriteHeader(http.StatusCreated)
	}
}

func setupDatabaseTables(db *sql.DB) {
	// Create the user table
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, username TEXT, email TEXT, password TEXT)")
	if err != nil {
		panic(err)
	}
}
