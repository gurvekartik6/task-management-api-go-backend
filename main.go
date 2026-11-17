package main

import (
	"fmt"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Home Page")
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `{"server" : "http://localhost:8080", "status" : "running"}`)
}
func main() {

	fmt.Println("This is the main function ")

	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/health", HealthHandler)

	fmt.Println("Server running on the http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Server Failed : ", err)
	}
}
