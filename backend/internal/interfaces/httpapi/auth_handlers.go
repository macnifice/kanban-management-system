package httpapi

import (
	"errors"
	"net/http"

	"kanban/internal/domain/identity"
	"kanban/internal/domain/shared"
)

type userDTO struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func toUserDTO(u *identity.User) userDTO { return userDTO{ID: u.ID, Email: u.Email, Name: u.Name} }

type credentials struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if !decode(r, &c) {
		fail(w, 400, "petición inválida")
		return
	}
	sess, err := s.Identity.Register(r.Context(), c.Email, c.Name, c.Password)
	if errors.Is(err, shared.ErrInvalid) {
		fail(w, 400, "email, nombre y contraseña (mín. 6 caracteres) son requeridos")
		return
	}
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"token": sess.Token, "user": toUserDTO(sess.User)})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if !decode(r, &c) {
		fail(w, 400, "petición inválida")
		return
	}
	sess, err := s.Identity.Login(r.Context(), c.Email, c.Password)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"token": sess.Token, "user": toUserDTO(sess.User)})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	u, err := s.Identity.Profile(r.Context(), userID(r))
	if errors.Is(err, shared.ErrNotFound) {
		fail(w, 401, "no autenticado")
		return
	}
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, 200, toUserDTO(u))
}

func (s *Server) users(w http.ResponseWriter, r *http.Request) {
	us, err := s.Identity.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	out := make([]userDTO, 0, len(us))
	for _, u := range us {
		out = append(out, toUserDTO(u))
	}
	writeJSON(w, 200, out)
}
