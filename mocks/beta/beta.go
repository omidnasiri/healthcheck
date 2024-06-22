package main

import (
	"log"
	"math/rand"
	"net/http"
)

func main() {

	http.HandleFunc("/beta", func(w http.ResponseWriter, r *http.Request) {
		chance := rand.Intn(100)
		if chance < 90 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	log.Println("beta server listening on port 8080")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalln(err)
	}
}
