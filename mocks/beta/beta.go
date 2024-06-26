package main

import (
	"log"
	"math/rand"
	"net/http"
)

func main() {

	http.HandleFunc("/beta", func(w http.ResponseWriter, r *http.Request) {
		chance := rand.Intn(100)
		var status int
		if chance < 90 {
			status = http.StatusOK
		} else {
			status = http.StatusInternalServerError
		}
		log.Println("beta server status:", status)
		w.WriteHeader(status)
	})

	log.Println("beta server listening on port 8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalln(err)
	}
}
