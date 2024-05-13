package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime/coverage"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leanovate/gopter/arbitrary"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/tools/cover"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

var users []User

func (u *User) IsEmpty() bool {
	return u.Username == ""
}

func SetupSwagger(router *mux.Router, swaggerEndPoint string) (*mux.Router, error) {
	// Ensure swaggerEndPoint is not empty
	if swaggerEndPoint == "" {

		return nil, errors.New("swaggerEndPoint cannot be empty")
	}
	// Ensure router is not nil
	if router == nil {

		return nil, errors.New("router cannot be empty")
	}

	// Serve Swagger UI
	router.PathPrefix("/swagger/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Serve Swagger UI using http.StripPrefix to handle the prefix
		httpSwagger.Handler(
			httpSwagger.URL(swaggerEndPoint), // URL pointing to API definition
		).ServeHTTP(w, r)
	}))

	// Serve Swagger JSON file
	router.Handle("/docs/swagger.json", http.FileServer(http.Dir("./")))

	return router, nil
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

	// Additional hooks for fuzzing?
	r.HandleFunc("/exit", exitProgram).Methods("GET")
	r.HandleFunc("/coverage", coverageSoFar).Methods("GET")
	r.HandleFunc("/generate", generateUser).Methods("GET")

	// Setup Swagger
	swaggerEndPoint := "/docs/swagger.json"
	router, err := SetupSwagger(r, swaggerEndPoint)
	if err != nil {
		log.Fatalf("Error setting up Swagger: %v", err)
	}
	// listen on port
	log.Fatal(http.ListenAndServe(":4000", router))
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

func exitProgram(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Calling Exit")
	os.Exit(0)
}

func coverageSoFar(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Called CoverageSoFar")
	covDir := uuid.NewString()
	if err := os.MkdirAll(fmt.Sprintf("./%s", covDir), os.ModePerm); err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	if err := coverage.WriteMetaDir(covDir); err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	if err := coverage.WriteCountersDir(covDir); err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	profileOutputFile := fmt.Sprintf("profile_%s.txt", covDir)
	covDataCmd := exec.Command("go", "tool", "covdata", "textfmt", "-i", covDir, "-o", profileOutputFile)
	if err := covDataCmd.Run(); err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	profiles, err := cover.ParseProfiles(profileOutputFile)
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	// cleanup the cov dir and profile file
	defer os.RemoveAll(covDir)
	defer os.Remove(profileOutputFile)

	coveredStmt := 0
	totalStmt := 0
	for _, profile := range profiles {
		for _, block := range profile.Blocks {
			if block.Count > 0 {
				coveredStmt += block.NumStmt
			}
			totalStmt += block.NumStmt
		}
	}

	result := make(map[string]any)
	result["count"] = coveredStmt
	result["stmt"] = totalStmt
	result["coverage"] = fmt.Sprintf("%.2f%%", (100 * float64(coveredStmt) / float64(totalStmt)))
	json.NewEncoder(w).Encode(result)
}

func generateUser(w http.ResponseWriter, r *http.Request) {
	arbitraries := arbitrary.DefaultArbitraries()
	var u User
	userGenerator := arbitraries.GenForType(reflect.TypeOf(u))
	sample, result := userGenerator.Sample()
	if !result {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode("Unable to generate a sample")
		return
	}
	json.NewEncoder(w).Encode(sample)
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
	if user.ID == "" {
		user.ID = strconv.Itoa(rand.Intn(100))
	}
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
func exitProgram(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Calling Exit")
	os.Exit(0)
}