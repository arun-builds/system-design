package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"e-commerce-catalog/internal/database"
)

const maxRequestBodyBytes = 1 << 20

func AddProductHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "application/json") {
			writeError(w, http.StatusUnsupportedMediaType, "content-type must be application/json")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

		var p database.Product
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}

		created, err := database.AddProduct(r.Context(), db, p)
		if err != nil {
			writeDBError(w, err, "add product")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func ListProductsHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		products, err := database.ListProducts(r.Context(), db)
		if err != nil {
			writeDBError(w, err, "list products")
			return
		}
		writeJSON(w, http.StatusOK, products)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeDBError(w http.ResponseWriter, err error, op string) {
	status, msg := mapProductError(err)
	if status >= 500 {
		log.Printf("%s: %v", op, err)
	}
	writeError(w, status, msg)
}

func mapProductError(err error) (int, string) {
	switch {
	case errors.Is(err, database.ErrNameRequired),
		errors.Is(err, database.ErrInvalidPrice),
		errors.Is(err, database.ErrInvalidStock):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
