package main

import (
	"log"
	"math/rand"
	"net/http"
)

func main() {

	http.HandleFunc("GET /alpha", func(w http.ResponseWriter, r *http.Request) {
		chance := rand.Intn(100)
		var status int
		if chance > 50 {
			status = http.StatusOK
		} else {
			status = http.StatusInternalServerError
		}
		log.Println("alpha server status:", status)
		w.WriteHeader(status)
	})

	log.Println("alpha server listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalln(err)
	}
}
