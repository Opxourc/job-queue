package main

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Timeout(TimeoutTime))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome"))
	})

	if err := http.ListenAndServe(":"+strconv.Itoa(AddressPort), r); err != nil {
		slog.Error(err.Error())
	}
}
