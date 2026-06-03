package main

import (
	"log"
	"net/http"

	"e-commerce-catalog/internal/api"
	"e-commerce-catalog/internal/database"
)

func main() {
	masterDSN := "postgres://arunprajapati@localhost:5432/postgres?sslmode=disable"
	replicaDSN := "postgres://arunprajapati@localhost:5433/postgres?sslmode=disable"

	db, err := database.ConnectDB(masterDSN, replicaDSN)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health-check", healthHandler)
	// mux.HandleFunc("/api/health-chec", healthHandler)

	// admin routes
	mux.HandleFunc("/api/admin", healthHandler)
	mux.HandleFunc("POST /api/admin/products", api.AddProductHandler(db))

	// public routes
	mux.HandleFunc("GET /api/products", api.ListProductsHandler(db))

	// user routes
	mux.HandleFunc("/api/user", healthHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("api listening on :8080")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
