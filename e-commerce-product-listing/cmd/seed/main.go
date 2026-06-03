package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"e-commerce-catalog/internal/database"
)

const totalProducts = 100

func main() {
	dsn := "postgres://arunprajapati@localhost:5432/postgres?sslmode=disable"
	if dsn == "" {
		log.Fatal("MASTER_DB_URL is required")
	}

	db, err := database.ConnectDB(dsn, dsn)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	var inserted, failed int

	for i := 1; i <= totalProducts; i++ {
		p := database.Product{
			Name:        fmt.Sprintf("Product %03d", i),
			Description: fmt.Sprintf("Auto-generated description for product number %d.", i),
			Price:       float64(rand.Intn(10000)) / 100,
			Stock:       rand.Intn(1000),
			ImageURL:    fmt.Sprintf("https://example.com/products/%03d.jpg", i),
		}

		if _, err := database.AddProduct(ctx, db, p); err != nil {
			log.Printf("insert %d: %v", i, err)
			failed++
			continue
		}
		inserted++
		if i%10 == 0 {
			log.Printf("progress: %d/%d", i, totalProducts)
		}
	}

	log.Printf("seed done: %d inserted, %d failed", inserted, failed)
}
