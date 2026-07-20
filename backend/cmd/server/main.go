package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/XiaoMoDern/dev-knowledge-base/backend/internal/httpapi"
	"github.com/XiaoMoDern/dev-knowledge-base/backend/internal/store"
)

const databasePath = "data/dev-notes.db"

func newServer(notesStore httpapi.NotesStore) *http.Server {
	return &http.Server{
		Addr:    "127.0.0.1:8181",
		Handler: httpapi.NewHandler(notesStore),
	}
}

func openStore() (*store.Store, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	database, err := store.Open(databasePath)
	if err != nil {
		return nil, fmt.Errorf("open application store: %w", err)
	}

	return database, nil
}

func main() {
	database, err := openStore()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	server := newServer(database)

	log.Printf("server listening on http://%s", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
