package main

import (
	"log"
	"net/http"

	"github.com/XiaoMoDern/dev-knowledge-base/backend/internal/httpapi"
)

func newServer() *http.Server {
	return &http.Server{
		Addr:    "127.0.0.1:8181",
		Handler: httpapi.NewHandler(),
	}
}

func main() {
	server := newServer()

	log.Printf("server listening on http://%s", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
