package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			log.Println("query failed:", err)
			return
		}
		fmt.Fprintf(w, "Users registered: %d\n", count)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
