package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"
	"encoding/json"

	_ "github.com/lib/pq"
)

var db *sql.DB

type Note struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}


func getNotes(w http.ResponseWriter, r *http.Request){
	rows, err := db.Query("SELECT id, body, created_at FROM notes ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer rows.Close()
	notes:=[]Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Body, &n.CreatedAt); err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		notes=append(notes, n)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}
func createNote(w http.ResponseWriter, r *http.Request) {
	var n Note
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		log.Println(err)
		return
	}
	if n.Body == "" {
		http.Error(w, "body is required", http.StatusBadRequest)
		return
	}
	err := db.QueryRow(
		"INSERT INTO notes (user_id,body) VALUES ($1, $2) RETURNING id, created_at",
		1, n.Body, ).Scan(&n.ID, &n.CreatedAt)

	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(n)

}

func main() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL not set")
	}

	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("open failed: ", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("cannot reach database: ", err)
	}
	log.Println("connected to database")


	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/notes", getNotes)
	mux.HandleFunc("POST /api/notes", createNote)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

