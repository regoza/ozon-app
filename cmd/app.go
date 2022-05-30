package main

import (
	"github.com/regoza/ozon-app/api"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/v1/signUp", api.SignUp)
	http.HandleFunc("/api/v1/signIn", api.SignIn)
	http.HandleFunc("/api/v1/products", api.Products)
	http.HandleFunc("/api/v1/logout", api.Logout)

	log.Fatal(http.ListenAndServe(":9001", nil))
}
