package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"maps"
	"net/http"
	"slices"

	_ "github.com/lib/pq"
	"github.com/mikemcavoydev/tables/migrations"
	"github.com/pressly/goose/v3"
)

//go:embed static/*
var static embed.FS

type Table struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Items []Item `json:"items"`
}

type Item struct {
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

	router.Handle("POST /api/tables", createTable(db))

	router.Handle("POST /api/tags", createTag(db))

	router.Handle("GET /api/tables", allTables(db))

	server := http.Server{
		Addr:    ":8080",
		Handler: corsMiddleware(router),
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func createTable(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dto struct {
			Title string `json:"title"`
		}

		err := json.NewDecoder(r.Body).Decode(&dto)
		if err != nil {
			log.Printf("ERROR: invalid payload: %v", err)
			WriteJSON(w, http.StatusBadRequest, Envelope{"error": "bad request"})
			return
		}

		query := `INSERT INTO tables (title) VALUES ($1) RETURNING id`
		var id int
		err = db.QueryRow(query, dto.Title).Scan(&id)
		if err != nil {
			log.Printf("ERROR: failed to insert new table: %v", err)
			WriteJSON(w, http.StatusInternalServerError, Envelope{"error": "something went wrong"})
			return
		}

		WriteJSON(w, http.StatusCreated, Envelope{"data": Table{
			ID:    id,
			Title: dto.Title,
		}})
	}
}

func createTag(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func allTables(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		tags := make(map[int][]Tag)
		query := `SELECT tags.id, tags.title, tags.description, item_id from tags
		INNER JOIN item_tags ON tag_id = tags.id;`

		itemTagsRows, err := db.Query(query)
		if err != nil {
			log.Printf("ERROR: failed to retrieve item tags: %v", err)
			WriteJSON(w, http.StatusInternalServerError, Envelope{"error": "something went wrong"})
			return
		}
		defer itemTagsRows.Close()

		for itemTagsRows.Next() {
			var tag Tag
			var item_id int
			itemTagsRows.Scan(&tag.ID, &tag.Title, &tag.Description, &item_id)

			tags[item_id] = append(tags[item_id], tag)
		}

		query = `select t.id, t.title, i.id, i.title FROM tables t
		LEFT JOIN items i ON i.table_id = t.id;`
		tableRows, err := db.Query(query)
		if err != nil {
			log.Printf("ERROR: failed to retrieve tables: %v", err)
			WriteJSON(w, http.StatusInternalServerError, Envelope{"error": "something went wrong"})
			return
		}
		defer tableRows.Close()

		tableMap := make(map[int]Table)
		for tableRows.Next() {
			var table Table
			var item Item
			tableRows.Scan(&table.ID, &table.Title, &item.ID, &item.Title)

			if itemTags, ok := tags[item.ID]; ok {
				item.Tags = itemTags
			}

			if existing, ok := tableMap[table.ID]; ok {
				existing.Items = append(existing.Items, item)
				tableMap[table.ID] = existing
			} else {
				table.Items = []Item{item}
				tableMap[table.ID] = table
			}

		}

		tables := slices.Collect(maps.Values(tableMap))

		WriteJSON(w, http.StatusCreated, Envelope{"data": tables})
	}
}
