package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	if _, err := os.Stat("./index.html"); os.IsNotExist(err) {
		log.Fatalf("index.html not found: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving %s to %s", r.URL.Path, r.RemoteAddr)
		http.ServeFile(w, r, "./index.html")
	})

	mux.HandleFunc("/assets/logo.png", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving %s to %s", r.URL.Path, r.RemoteAddr)
		http.ServeFile(w, r, "./assets/logo.png")
	})

	srv := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	fmt.Println("Starting server on :8080")
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Server start error: %v", err)
	}
}
