package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func main() {

	http.HandleFunc("/webhook/{id}", func(w http.ResponseWriter, r *http.Request) {
		payload := struct {
			Status bool `json:"status"`
		}{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&payload)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Println("webhook called, id: ", strings.TrimPrefix(r.URL.Path, "/webhook/"), "status:", payload.Status)
		w.WriteHeader(http.StatusOK)
	})

	log.Println("webhook listening on port 8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalln(err)
	}
}
