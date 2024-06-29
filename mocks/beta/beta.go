package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
)

func main() {

	http.HandleFunc("POST /beta", func(w http.ResponseWriter, r *http.Request) {
		payload := struct {
			Ping string `json:"ping"`
		}{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&payload)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error: "invalid payload}`))
			return
		}

		chance := rand.Intn(100)
		var status int
		if chance < 90 {
			status = http.StatusOK
		} else {
			status = http.StatusInternalServerError
		}

		log.Println("beta server status:", status)
		w.WriteHeader(status)
		w.Write([]byte(`{"pong":"pong"}`))
	})

	log.Println("beta server listening on port 8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalln(err)
	}
}
