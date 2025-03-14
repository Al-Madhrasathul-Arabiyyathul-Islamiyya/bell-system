package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, server is up!"))
	})

	http.ListenAndServe(":8080", r)
}
