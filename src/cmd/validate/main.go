package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := ":8081"
	if v := os.Getenv("LISTEN_ADDR"); v != "" {
		addr = v
	}
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	http.HandleFunc("/v1/validate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "json", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	log.Printf("validate service %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
