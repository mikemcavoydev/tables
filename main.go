package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/mikemcavoydev/tables/migrations"
	"github.com/pressly/goose/v3"
)

//go:embed static/*
var static embed.FS

type Items struct {
	ID      int     `json:"id"`
	Title   string  `json:"title"`
	Entries []Entry `json:"entries"`
}

type Entry struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Tags  []Tag  `json:"tags"`
}

type Tag struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func main() {
	fmt.Println("starting server...")

	db := dbInit()
	defer db.Close()

	router := http.NewServeMux()

	router.Handle("/", http.FileServer(http.FS(getStaticFS())))

	router.HandleFunc("POST /api/tables", func(w http.ResponseWriter, r *http.Request) {
		var dto struct {
			Title string `json:"title"`
		}

		err := json.NewDecoder(r.Body).Decode(&dto)
		if err != nil {
			log.Printf("ERROR: invalid payload: %v", err)
			WriteJSON(w, http.StatusBadRequest, Envelope{"error": "bad request"})
			return
		}

		query := `INSERT INTO items (title) VALUES ($1) RETURNING id`
		var id int
		err = db.QueryRow(query, dto.Title).Scan(&id)
		if err != nil {
			log.Printf("ERROR: failed to insert new items: %v", err)
			WriteJSON(w, http.StatusInternalServerError, Envelope{"error": "something went wrong"})
			return
		}

		WriteJSON(w, http.StatusCreated, Envelope{"data": Items{
			ID:    id,
			Title: dto.Title,
		}})
	})

	router.HandleFunc("POST /api/tags", func(w http.ResponseWriter, r *http.Request) {
		var dto struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}

		err := json.NewDecoder(r.Body).Decode(&dto)
		if err != nil {
			log.Printf("ERROR: invalid payload: %v", err)
			WriteJSON(w, http.StatusBadRequest, Envelope{"error": "bad request"})
			return
		}

		query := `INSERT INTO tags (title, description) VALUES ($1, $2) RETURNING id`
		var id int
		err = db.QueryRow(query, dto.Title, dto.Description).Scan(&id)
		if err != nil {
			log.Printf("ERROR: failed to insert new tag: %v", err)
			WriteJSON(w, http.StatusInternalServerError, Envelope{"error": "something went wrong"})
			return
		}

		WriteJSON(w, http.StatusCreated, Envelope{"data": Tag{
			ID:          id,
			Title:       dto.Title,
			Description: dto.Description,
		}})
	})

	router.HandleFunc("POST /api/tables/{id}", func(w http.ResponseWriter, r *http.Request) {
		tableId, err := strconv.ParseInt(r.PathValue("id"), 0, 64)
		if err != nil {
			log.Printf("ERROR: invalid id: %v", err)
			WriteJSON(w, http.StatusBadRequest, Envelope{"error": "bad request"})
			return
		}

		var dto struct {
			Title string `json:"title"`
			Tags  []int  `json:"tags"`
		}

		err = json.NewDecoder(r.Body).Decode(&dto)
		if err != nil {
			log.Printf("ERROR: invalid payload: %v", err)
			WriteJSON(w, http.StatusBadRequest, Envelope{"error": "bad request"})
			return
		}

		query := `INSERT INTO entries (title, item_id) VALUES ($1, $2) RETURNING id`
		var entry_id int
		err = db.QueryRow(query, dto.Title, tableId).Scan(&entry_id)
		if err != nil {
			log.Printf("ERROR: failed to insert new items: %v", err)
			WriteJSON(w, http.StatusInternalServerError, Envelope{"error": "something went wrong"})
			return
		}

		var tags []Tag
		for _, tag_id := range dto.Tags {
			query := `INSERT INTO entry_tags (tag_id, entry_id) VALUES ($1, $2)`
			_, err = db.Exec(query, tag_id, entry_id)
			if err != nil {
				log.Printf("ERROR: failed to insert new entry tags: %v", err)
				WriteJSON(w, http.StatusInternalServerError, Envelope{"error": "something went wrong"})
				return
			}

			query = `SELECT (id, title, description) FROM tags WHERE id = $1`
			err = db.QueryRow(query, tag_id).Scan(&entry_id)
			if err != nil {
				log.Printf("ERROR: failed to retrieve tags: %v", err)
				WriteJSON(w, http.StatusInternalServerError, Envelope{"error": "something went wrong"})
				return
			}
		}

		WriteJSON(w, http.StatusCreated, Envelope{"data": Entry{
			ID:    entry_id,
			Title: dto.Title,
			Tags:  tags,
		}})
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

func dbInit() *sql.DB {
	connStr := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("ERROR: could not connect to db: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("ERROR: db healthcheck failed: %v", err)
	}

	err = applyMigrations(db)
	if err != nil {
		log.Fatalf("ERROR: migrations were not applied successfully: %v", err)
	}

	return db
}

func applyMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrations.FS)
	defer func() {
		goose.SetBaseFS(nil)
	}()

	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	err = goose.Up(db, ".")
	if err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

type Envelope map[string]interface{}

func WriteJSON(w http.ResponseWriter, status int, data Envelope) error {
	js, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}
