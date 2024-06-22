package main

import (
	"log"
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

	log.Println("alpha server listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalln(err)
	}
}
