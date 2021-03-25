package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"encoding/json"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// StatusWriter to log response status code
type StatusWriter struct {
	http.ResponseWriter
	Status int
}

// NewStatusResponseWriter to create a new StatusWriter handler for HTTP logging
func NewStatusResponseWriter(w http.ResponseWriter) *StatusWriter {
	return &StatusWriter{
		ResponseWriter: w,
		Status:         http.StatusOK,
	}
}

// WriteHeader to write status code to response
func (sw *StatusWriter) WriteHeader(code int) {
	sw.Status = code
	sw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware to log HTTP requests
func LoggingMiddleware(r *mux.Router) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			sw := NewStatusResponseWriter(w)

			defer func() {
				log.Printf("%s %s %v %d %s %s", req.RemoteAddr, req.Method, time.Since(start), sw.Status, req.URL.Path, req.URL.RawQuery)
			}()
			next.ServeHTTP(w, req)
		})
	}
}

// handle health checks
func handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

// shopsHandler to expose shops over HTTP with a database connection
type shopsHandler struct {
	db *gorm.DB
}

// ServeHTTP to implement the handle interface for serving shops
func (h *shopsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var shops []Shop
	trx := h.db.Find(&shops)
	if trx.Error == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(shops)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// shopHandler to expose shops over HTTP with a database connection
type shopHandler struct {
	db *gorm.DB
}

// ServeHTTP to implement the handle interface for serving shop
func (h *shopHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]

	var shop Shop
	trx := h.db.First(&shop, id)
	if trx.Error == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(shop)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// productsHandler to expose products over HTTP with a database connection
type productsHandler struct {
	db *gorm.DB
}

// ServeHTTP to implement the handle interface for serving products
func (h *productsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var products []Product
	var trx *gorm.DB
	availableFilter := r.URL.Query().Get("available")

	if availableFilter != "" {
		available, err := strconv.ParseBool(availableFilter)
		if err != nil {
			log.Warnf("cannot parse available query to boolean: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			trx = h.db.Preload("Shop").Where(map[string]interface{}{"available": available}).Find(&products)
		}
	} else {
		trx = h.db.Preload("Shop").Find(&products)
	}

	if trx.Error == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(products)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// productHandler to expose product over HTTP with a database connection
type productHandler struct {
	db *gorm.DB
}

// ServeHTTP to implement the handle interface for serving product
func (h *productHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var product Product
	trx := h.db.Preload("Shop").First(&product, id)
	if trx.Error == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// StartAPI to handle HTTP requests
func StartAPI(db *gorm.DB, config ApiConfig) error {
	router := mux.NewRouter().StrictSlash(true)

	router.Path("/health").HandlerFunc(handleHealth)

	router.Path("/shops").Handler(&shopsHandler{db: db})
	router.Path("/shops/{id:[0-9]+}").Handler(&shopHandler{db: db})

	router.Path("/products").Handler(&productsHandler{db: db})
	router.Path("/products/{id:[0-9]+}").Handler(&productHandler{db: db})

	// register middlewares
	router.Use(LoggingMiddleware(router))

	log.Printf("starting API on %s", config.Address)
	if config.Certfile != "" && config.Keyfile != "" {
		return http.ListenAndServeTLS(config.Address, config.Certfile, config.Keyfile, router)
	} else {
		return http.ListenAndServe(config.Address, router)
	}
}
