// Package identity es el contexto de usuarios y autenticación.
package identity

import (
	"context"
	"errors"
	"strings"

	"kanban/internal/domain/shared"
)

const MinPasswordLength = 6

var (
	ErrEmailTaken         = errors.New("el email ya está registrado")
	ErrInvalidCredentials = errors.New("credenciales inválidas")
)

// User es la raíz del agregado de usuario.
type User struct {
	ID    int
	Email string
	Name  string
}

// NewUser valida los datos de registro. La contraseña solo se valida aquí; nunca se guarda en la entidad.
func NewUser(email, name, password string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name, err := shared.RequiredText(name, 120)
	if err != nil || !strings.Contains(email, "@") || len(email) > 190 || len(password) < MinPasswordLength {
		return nil, shared.ErrInvalid
	}
	return &User{Email: email, Name: name}, nil
}

// Puertos que implementa la infraestructura.

type UserRepository interface {
	// Add persiste el usuario y le asigna ID. Devuelve ErrEmailTaken si el email existe.
	Add(ctx context.Context, u *User, passwordHash string) error
	ByEmail(ctx context.Context, email string) (*User, string, error)
	ByID(ctx context.Context, id int) (*User, error)
	List(ctx context.Context) ([]*User, error)
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Matches(hash, plain string) bool
}

type TokenIssuer interface {
	Issue(userID int) (string, error)
	Verify(token string) (userID int, err error)
}
