package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"brand-protection-monitor/backend/internal/certificates"
	"brand-protection-monitor/backend/internal/exporter"
	"brand-protection-monitor/backend/internal/keywords"
	"brand-protection-monitor/backend/internal/monitor"

	"github.com/gorilla/mux"
)

type Server struct {
	keywordRepo     *keywords.Repository
	certificateRepo *certificates.Repository
	stateRepo       *monitor.StateRepository
	monitorService  *monitor.Service
	allowedOrigins  map[string]struct{}
}

func NewServer(keywordRepo *keywords.Repository, certificateRepo *certificates.Repository, stateRepo *monitor.StateRepository, monitorService *monitor.Service, allowedOrigins []string) *Server {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		originSet[origin] = struct{}{}
	}

	return &Server{
		keywordRepo:     keywordRepo,
		certificateRepo: certificateRepo,
		stateRepo:       stateRepo,
		monitorService:  monitorService,
		allowedOrigins:  originSet,
	}
}

func (s *Server) Router() http.Handler {
	router := mux.NewRouter()
	router.Use(s.corsMiddleware)

	router.HandleFunc("/health", s.health).Methods(http.MethodGet)
	router.HandleFunc("/api/keywords", s.listKeywords).Methods(http.MethodGet)
	router.HandleFunc("/api/keywords", s.createKeyword).Methods(http.MethodPost)
	router.HandleFunc("/api/keywords/{id}", s.deleteKeyword).Methods(http.MethodDelete)
	router.HandleFunc("/api/matches", s.listMatches).Methods(http.MethodGet)
	router.HandleFunc("/api/status", s.getStatus).Methods(http.MethodGet)
	router.HandleFunc("/api/export.csv", s.exportCSV).Methods(http.MethodGet)
	router.HandleFunc("/api/monitor/run-once", s.runMonitorOnce).Methods(http.MethodPost)
	router.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	return router
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) listKeywords(w http.ResponseWriter, r *http.Request) {
	items, err := s.keywordRepo.List(r.Context())
	if err != nil {
		writeInternalError(w, "failed to list keywords", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

type createKeywordRequest struct {
	Value string `json:"value"`
}

func (s *Server) createKeyword(w http.ResponseWriter, r *http.Request) {
	var request createKeywordRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if request.Value == "" {
		writeError(w, http.StatusBadRequest, "keyword value is required")
		return
	}

	item, err := s.keywordRepo.Create(r.Context(), request.Value)
	if err != nil {
		writeInternalError(w, "failed to create keyword", err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) deleteKeyword(w http.ResponseWriter, r *http.Request) {
	idValue := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid keyword id")
		return
	}

	if err := s.keywordRepo.Delete(r.Context(), id); err != nil {
		writeInternalError(w, "failed to delete keyword", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listMatches(w http.ResponseWriter, r *http.Request) {
	items, err := s.certificateRepo.List(r.Context())
	if err != nil {
		writeInternalError(w, "failed to list matches", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) getStatus(w http.ResponseWriter, r *http.Request) {
	state, err := s.stateRepo.Get(r.Context())
	if err != nil {
		writeInternalError(w, "failed to get monitor status", err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (s *Server) runMonitorOnce(w http.ResponseWriter, r *http.Request) {
	if err := s.monitorService.RunOnce(r.Context()); err != nil {
		if errors.Is(err, monitor.ErrRunAlreadyInProgress) {
			writeError(w, http.StatusConflict, "monitor run already in progress")
			return
		}
		writeInternalError(w, "failed to run monitor", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (s *Server) exportCSV(w http.ResponseWriter, r *http.Request) {
	items, err := s.certificateRepo.List(r.Context())
	if err != nil {
		writeInternalError(w, "failed to export matches", err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="matched-certificates.csv"`)

	if err := exporter.WriteMatchesCSV(w, items); err != nil {
		writeInternalError(w, "failed to write CSV export", err)
		return
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeInternalError(w http.ResponseWriter, clientMessage string, internalErr error) {
	log.Printf("%s: %v", clientMessage, internalErr)
	writeError(w, http.StatusInternalServerError, clientMessage)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if !s.isAllowedOrigin(origin) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) isAllowedOrigin(origin string) bool {
	_, ok := s.allowedOrigins[origin]
	return ok
}
