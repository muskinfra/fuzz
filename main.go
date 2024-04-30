package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

var users []User

func (u *User) IsEmpty() bool {
	return u.Username == ""
}
// @title User Management API
// @description This is a simple API for managing users
// @basePath /api/v1
func main() {
	fmt.Println("User Management API")
	r := mux.NewRouter()

	// seeding
	users = append(users, User{ID: "1", Username: "user1"})
	users = append(users, User{ID: "2", Username: "user2"})

	// routing
	r.HandleFunc("/", serveHome).Methods("GET")
	r.HandleFunc("/users", getAllUsers).Methods("GET")
	r.HandleFunc("/user/{id}", getUser).Methods("GET")
	r.HandleFunc("/user", createUser).Methods("POST")
	r.HandleFunc("/user/{id}", updateUser).Methods("PUT")
	r.HandleFunc("/user/{id}", deleteUser).Methods("DELETE")
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// listen on port
	log.Fatal(http.ListenAndServe(":4000", r))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>Welcome to User Management API</h1>"))
}
// @summary Get all users
// @description Get a list of all users
// @produce json
// @success 200 {array} User
// @router /users [get]
func getAllUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get all users")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
// @summary Get one user
// @description Get details of a single user by ID
// @produce json
// @param id path string true "User ID"
// @success 200 {object} User
// @failure 404 {string} string
// @router /user/{id} [get]
func getUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get one user")
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, user := range users {
		if user.ID == params["id"] {
			json.NewEncoder(w).Encode(user)
			return
		}
	}
	json.NewEncoder(w).Encode("No user found with given id")
}
// @summary Create one user
// @description Create a new user
// @accept json
// @produce json
// @param body body User true "User details"
// @success 200 {object} User
// @failure 400 {string} string
// @router /user [post]
func createUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create one user")
	w.Header().Set("Content-Type", "application/json")
	if r.Body == nil {
		json.NewEncoder(w).Encode("Please send some data")
	}
	var user User
	_ = json.NewDecoder(r.Body).Decode(&user)
	if user.IsEmpty() {
		json.NewEncoder(w).Encode("No data inside JSON")
		return
	}
	rand.Seed(time.Now().UnixNano())
	user.ID = strconv.Itoa(rand.Intn(100))
	users = append(users, user)
	json.NewEncoder(w).Encode(user)
}
// @summary Update one user
// @description Update details of an existing user
// @accept json
// @produce json
// @param id path string true "User ID"
// @param body body User true "Updated user details"
// @success 200 {object} User
// @failure 404 {string} string
// @router /user/{id} [put]
func updateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update one user")
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, user := range users {
		if user.ID == params["id"] {
			users = append(users[:index], users[index+1:]...)
			var user User
			_ = json.NewDecoder(r.Body).Decode(&user)
			user.ID = params["id"]
			users = append(users, user)
			json.NewEncoder(w).Encode(user)
			return
		}
	}
}
// @summary Delete one user
// @description Delete an existing user
// @produce json
// @param id path string true "User ID"
// @success 200 {string} string
// @failure 404 {string} string
// @router /user/{id} [delete]
func deleteUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete one user")
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, user := range users {
		if user.ID == params["id"] {
			users = append(users[:index], users[index+1:]...)
			json.NewEncoder(w).Encode("User deleted successfully")
			return
		}
	}
}
