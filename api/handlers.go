package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type session struct {
	Login  string
	Expiry time.Time
}

type product struct {
	Name  string
	Price float64
}

var usersDb = make(map[string]string)

var sessions = map[string]session{}

var products = []product{
	{"Kayak", 279},
	{"Life-Jacket", 49.95},
	{"Soccer Ball", 19.50},
	{"Hockey stick", 34.95},
	{"Hockey puck", 12},
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, exist := usersDb[creds.Email]; exist {
		// if user with the same email already exists
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("sorry, but user with the same email already exists"))
		return
	}

	usersDb[creds.Email] = creds.Password
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("user with email %s was saccefully registred!, please sign in then", creds.Email)))
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the expected password from our inmemory db
	expectedPassword, ok := usersDb[creds.Email]

	if !ok || expectedPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Create a new random session token
	sessionToken := uuid.NewString()
	loc, _ := time.LoadLocation("Europe/Moscow")
	expiresAt := time.Now().In(loc).Add(120 * time.Second)

	// Set the token in the session map, along with the user whom it represents
	sessions[sessionToken] = session{
		Login:  creds.Email,
		Expiry: expiresAt,
	}

	// we also set an expiry time of 120 seconds for our cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})
}

func Products(w http.ResponseWriter, r *http.Request) {
	// try to get cookies from request
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	// We then get the name of the user from our session map, where we set the session token
	userSession, exists := sessions[sessionToken]
	if !exists {
		// If the session token is not present in session map, return an unauthorized error
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if userSession.isExpired() {
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err = w.Write([]byte(fmt.Sprintf("Products: %v", products)))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	// remove the user session from the session map
	delete(sessions, sessionToken)

	// clean cookies
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now(),
	})
}

func (s session) isExpired() bool {
	return s.Expiry.Before(time.Now())
}
