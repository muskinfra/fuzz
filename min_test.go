package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHome(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(serveHome)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "<h1>Welcome to User Management API</h1>"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestGetAllUsers(t *testing.T) {
	req, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getAllUsers)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var users []User
	err = json.Unmarshal(rr.Body.Bytes(), &users)
	if err != nil {
		t.Errorf("error decoding response body: %v", err)
	}

	
}
func TestCreateUser(t *testing.T) {
	// Create a request body with user data
	userData := map[string]string{
		"username": "testuser",
	}
	reqBody, _ := json.Marshal(userData)

	req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createUser)

	// Serve the HTTP request
	handler.ServeHTTP(rr, req)

	// Check the HTTP response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	
	var createdUser User
	err = json.Unmarshal(rr.Body.Bytes(), &createdUser)
	if err != nil {
		t.Errorf("error decoding response body: %v", err)
	}

	
	if createdUser.ID == "" {
		t.Error("user ID is empty")
	}

	
	if createdUser.Username != userData["username"] {
		t.Errorf("username mismatch: got %s want %s", createdUser.Username, userData["username"])
	}

	
}

