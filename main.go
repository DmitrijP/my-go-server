package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	srv := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	err := srv.ListenAndServe()
	if err != nil {
		fmt.Errorf("Start Error %w", err)
	}
}
