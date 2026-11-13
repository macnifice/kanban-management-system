// Package httpapi es la capa de interfaz HTTP: traduce peticiones a casos de uso y errores de dominio a códigos HTTP.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	appid "kanban/internal/application/identity"
	appkb "kanban/internal/application/kanban"
	"kanban/internal/domain/identity"
	"kanban/internal/domain/kanban"
	"kanban/internal/domain/shared"
)

type Server struct {
	Identity       *appid.Service
	Tokens         identity.TokenIssuer
	Boards         *appkb.BoardService
	Cards          *appkb.CardService
	Labels         *appkb.LabelService
	Queries        *appkb.QueryService
	Ping           func(ctx context.Context) error
	AllowedOrigins []string
}

type ctxKey struct{}

func (s *Server) Routes() http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("GET /api/me", s.me)
	api.HandleFunc("GET /api/users", s.users)

	api.HandleFunc("GET /api/boards", s.listBoards)
	api.HandleFunc("POST /api/boards", s.createBoard)
	api.HandleFunc("GET /api/boards/{id}", s.getBoard)
	api.HandleFunc("PATCH /api/boards/{id}", s.updateBoard)
	api.HandleFunc("DELETE /api/boards/{id}", s.deleteBoard)
	api.HandleFunc("GET /api/boards/{id}/metrics", s.metrics)
	api.HandleFunc("POST /api/boards/{id}/columns", s.addColumn)
	api.HandleFunc("PUT /api/boards/{id}/columns/order", s.reorderColumns)
	api.HandleFunc("POST /api/boards/{id}/labels", s.createLabel)

	api.HandleFunc("PATCH /api/columns/{id}", s.updateColumn)
	api.HandleFunc("DELETE /api/columns/{id}", s.deleteColumn)

	api.HandleFunc("PATCH /api/labels/{id}", s.updateLabel)
	api.HandleFunc("DELETE /api/labels/{id}", s.deleteLabel)

	api.HandleFunc("POST /api/cards", s.createCard)
	api.HandleFunc("PATCH /api/cards/{id}", s.updateCard)
	api.HandleFunc("DELETE /api/cards/{id}", s.deleteCard)
	api.HandleFunc("PUT /api/cards/{id}/move", s.moveCard)
	api.HandleFunc("GET /api/cards/{id}/history", s.history)

	root := http.NewServeMux()
	root.HandleFunc("GET /api/health", s.health)
	root.HandleFunc("POST /api/auth/register", s.register)
	root.HandleFunc("POST /api/auth/login", s.login)
	root.Handle("/api/", s.authenticate(api))
	return logging(s.cors(root))
}

// ---- middleware ----

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		tok := strings.TrimPrefix(h, "Bearer ")
		if tok == "" || tok == h {
			fail(w, http.StatusUnauthorized, "no autenticado")
			return
		}
		id, err := s.Tokens.Verify(tok)
		if err != nil {
			fail(w, http.StatusUnauthorized, "no autenticado")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
	})
}

func userID(r *http.Request) int {
	id, _ := r.Context().Value(ctxKey{}).(int)
	return id
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, o := range s.AllowedOrigins {
			if o == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// ---- respuestas y errores ----

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

func fail(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// failErr traduce errores de dominio a respuestas HTTP.
func failErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		fail(w, http.StatusNotFound, "no encontrado")
	case errors.Is(err, shared.ErrInvalid):
		fail(w, http.StatusBadRequest, "petición inválida")
	case errors.Is(err, kanban.ErrWIPLimitReached):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error(), "code": "wip_limit"})
	case errors.Is(err, identity.ErrEmailTaken):
		fail(w, http.StatusConflict, err.Error())
	case errors.Is(err, identity.ErrInvalidCredentials):
		fail(w, http.StatusUnauthorized, err.Error())
	default:
		log.Printf("error: %v", err)
		fail(w, http.StatusInternalServerError, "error interno")
	}
}

func decode(r *http.Request, v any) bool { return json.NewDecoder(r.Body).Decode(v) == nil }

func decodeMap(r *http.Request) (map[string]any, bool) {
	var m map[string]any
	if json.NewDecoder(r.Body).Decode(&m) != nil || m == nil {
		return nil, false
	}
	return m, true
}

func pathID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	return id, err == nil && id > 0
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.Ping(r.Context()); err != nil {
		fail(w, http.StatusServiceUnavailable, "db no disponible")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func ok(w http.ResponseWriter) { writeJSON(w, 200, map[string]string{"status": "ok"}) }
