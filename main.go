package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

//go:embed static/*
var static embed.FS

type Table struct {
	ID      int64        `json:"id"`
	Title   string       `json:"title"`
	Entries []TableEntry `json:"entries"`
}

type TableEntry struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Tags  []Tag  `json:"tags"`
}

type Tag struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func main() {
	fmt.Println("starting server...")

	router := http.NewServeMux()

	router.Handle("/", http.FileServer(http.FS(getStaticFS())))

	router.HandleFunc("GET /api/tables/{id}", func(w http.ResponseWriter, r *http.Request) {
		table := &Table{
			ID:    1,
			Title: "Games Completion List",
			Entries: []TableEntry{
				{
					ID: 1,
					Tags: []Tag{
						{
							ID:          1,
							Title:       "In Progress",
							Description: "Group games that are currently in progress",
						},
					},
				},
			},
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(table)
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	server.ListenAndServe()
}

func getStaticFS() fs.FS {
	_, err := fs.Stat(static, "static")
	if err != nil {
		log.Fatalf("ERROR: could not read static directory: %v", err)
	}

	files, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatal("ERROR: invalid path provided")
	}

	return files
}
