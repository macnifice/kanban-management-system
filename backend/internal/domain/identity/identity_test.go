package identity

import (
	"errors"
	"testing"

	"kanban/internal/domain/shared"
)

func TestNewUser(t *testing.T) {
	u, err := NewUser("  Ana@Mail.COM ", " Ana ", "secret1")
	if err != nil || u.Email != "ana@mail.com" || u.Name != "Ana" {
		t.Fatalf("usuario inesperado: %+v, %v", u, err)
	}
	for _, c := range []struct{ e, n, p string }{
		{"sin-arroba", "Ana", "secret1"},
		{"a@b.c", "", "secret1"},
		{"a@b.c", "Ana", "corta"},
	} {
		if _, err := NewUser(c.e, c.n, c.p); !errors.Is(err, shared.ErrInvalid) {
			t.Fatalf("%+v debería ser inválido", c)
		}
	}
}
