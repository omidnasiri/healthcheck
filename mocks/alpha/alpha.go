package main

import (
	"math/rand"
	"net/http"
)

func main() {

	http.HandleFunc("/alpha", func(w http.ResponseWriter, r *http.Request) {
		chance := rand.Intn(100)
		if chance < 90 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	http.ListenAndServe(":8080", nil)
}
