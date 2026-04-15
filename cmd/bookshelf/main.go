package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/LuhTonkaYeat/bookshelf/proto"
)

type SaveQuoteRequest struct {
	Quote  string `json:"quote"`
	Author string `json:"author"`
}

type SaveQuoteResponse struct {
	ID     int    `json:"id"`
	Quote  string `json:"quote"`
	Author string `json:"author"`
	Saved  bool   `json:"saved"`
}

var (
	db         *sql.DB
	grpcClient pb.AuthorValidatorClient
)

func main() {
	initDB()
	defer db.Close()

	initGRPC()

	r := mux.NewRouter()
	r.HandleFunc("/save", saveQuoteHandler).Methods("POST")
	r.HandleFunc("/health", healthHandler).Methods("GET")

	port := ":8080"
	log.Printf("Bookshelf service running on %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}

func initGRPC() {
	addr := os.Getenv("VALIDATOR_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("failed to connect to validator: %v", err)
	}

	grpcClient = pb.NewAuthorValidatorClient(conn)
	log.Printf("Connected to validator gRPC service at %s", addr)
}

func initDB() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	connStr := fmt.Sprintf("postgres://postgres:postgres@%s:5432/bookshelf?sslmode=disable", host)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Println("Connected to PostgreSQL")

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS quotes (
        id SERIAL PRIMARY KEY,
        quote TEXT NOT NULL,
        author VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("failed to create table: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func saveQuoteHandler(w http.ResponseWriter, r *http.Request) {
	var req SaveQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Quote == "" || req.Author == "" {
		http.Error(w, "Quote and author are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcResp, err := grpcClient.Validate(ctx, &pb.ValidateRequest{
		Author: req.Author,
	})

	if err != nil {
		log.Printf("gRPC error: %v", err)
		http.Error(w, "Validator service unavailable", http.StatusServiceUnavailable)
		return
	}

	if !grpcResp.Valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":  "author is banned",
			"reason": grpcResp.Reason,
		})
		return
	}

	var id int
	err = db.QueryRow(
		"INSERT INTO quotes (quote, author) VALUES ($1, $2) RETURNING id",
		req.Quote, req.Author,
	).Scan(&id)

	if err != nil {
		log.Printf("DB error: %v", err)
		http.Error(w, "Failed to save quote", http.StatusInternalServerError)
		return
	}

	response := SaveQuoteResponse{
		ID:     id,
		Quote:  req.Quote,
		Author: req.Author,
		Saved:  true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
