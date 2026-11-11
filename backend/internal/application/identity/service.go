// Package identity (aplicación) contiene los casos de uso de registro, acceso y consulta de usuarios.
package identity

import (
	"context"
	"strings"

	domain "kanban/internal/domain/identity"
)

type Service struct {
	users  domain.UserRepository
	hasher domain.PasswordHasher
	tokens domain.TokenIssuer
}

func NewService(users domain.UserRepository, hasher domain.PasswordHasher, tokens domain.TokenIssuer) *Service {
	return &Service{users: users, hasher: hasher, tokens: tokens}
}

type Session struct {
	Token string
	User  *domain.User
}

func (s *Service) Register(ctx context.Context, email, name, password string) (*Session, error) {
	u, err := domain.NewUser(email, name, password)
	if err != nil {
		return nil, err
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}
	if err := s.users.Add(ctx, u, hash); err != nil {
		return nil, err
	}
	return s.session(u)
}

func (s *Service) Login(ctx context.Context, email, password string) (*Session, error) {
	u, hash, err := s.users.ByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil || !s.hasher.Matches(hash, password) {
		return nil, domain.ErrInvalidCredentials
	}
	return s.session(u)
}

func (s *Service) session(u *domain.User) (*Session, error) {
	tok, err := s.tokens.Issue(u.ID)
	if err != nil {
		return nil, err
	}
	return &Session{Token: tok, User: u}, nil
}

func (s *Service) Profile(ctx context.Context, id int) (*domain.User, error) {
	return s.users.ByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*domain.User, error) {
	return s.users.List(ctx)
}
