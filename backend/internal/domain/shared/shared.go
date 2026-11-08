// Package shared contiene el núcleo compartido por los contextos del dominio.
package shared

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrNotFound = errors.New("no encontrado")
	ErrInvalid  = errors.New("petición inválida")
)

// RequiredText recorta s y valida que no esté vacío ni supere max caracteres.
func RequiredText(s string, max int) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > max {
		return "", ErrInvalid
	}
	return s, nil
}
