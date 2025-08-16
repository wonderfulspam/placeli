package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/user/placeli/internal/database"
	"github.com/user/placeli/internal/logger"
	"github.com/user/placeli/internal/models"
)

//go:embed templates/*
var templates embed.FS

//go:embed static/*
var static embed.FS

type Server struct {
	db       *database.DB
	tmpl     *template.Template
	port     int
	apiKey   string
}

func NewServer(db *database.DB, port int, apiKey string) (*Server, error) {
	tmpl, err := template.ParseFS(templates, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Server{
		db:     db,
		tmpl:   tmpl,
		port:   port,
		apiKey: apiKey,
	}, nil
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/places", s.handleAPIPlaces)
	mux.HandleFunc("/api/place/", s.handleAPIPlace)
	mux.Handle("/static/", http.FileServer(http.FS(static)))

	addr := fmt.Sprintf(":%d", s.port)
	logger.Info("Starting web server", "address", addr)
	fmt.Printf("Web interface available at http://localhost:%d\n", s.port)
	
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := struct {
		Title  string
		APIKey string
	}{
		Title:  "Placeli - Saved Places",
		APIKey: s.apiKey,
	}

	if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		logger.Error("Failed to render template", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (s *Server) handleAPIPlaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")
	search := query.Get("search")

	limit := 100
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var places []*models.Place
	var err error

	if search != "" {
		places, err = s.db.SearchPlaces(search)
	} else {
		places, err = s.db.ListPlaces(limit, offset)
	}

	if err != nil {
		logger.Error("Failed to fetch places", "error", err)
		http.Error(w, "Failed to fetch places", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(places); err != nil {
		logger.Error("Failed to encode response", "error", err)
	}
}

func (s *Server) handleAPIPlace(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/place/"):]
	if id == "" {
		http.Error(w, "Place ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		place, err := s.db.GetPlace(id)
		if err != nil {
			http.Error(w, "Place not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(place); err != nil {
			logger.Error("Failed to encode response", "error", err)
		}

	case http.MethodPut:
		var place models.Place
		if err := json.NewDecoder(r.Body).Decode(&place); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		place.ID = id
		if err := s.db.SavePlace(&place); err != nil {
			logger.Error("Failed to update place", "error", err)
			http.Error(w, "Failed to update place", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(&place); err != nil {
			logger.Error("Failed to encode response", "error", err)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}