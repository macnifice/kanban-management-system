// Package kanban es el contexto principal: tableros, columnas con límite de WIP y tarjetas.
package kanban

import (
	"errors"

	"kanban/internal/domain/shared"
)

var (
	// ErrWIPLimitReached se devuelve al intentar meter una tarjeta en una columna que alcanzó su límite.
	ErrWIPLimitReached = errors.New("la columna alcanzó su límite de WIP")
	// ErrConcurrentChange indica que otro proceso modificó el agregado; la operación puede reintentarse.
	ErrConcurrentChange = errors.New("cambio concurrente, reintentar")
)

// Priority es un objeto de valor.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

func ParsePriority(s string) (Priority, error) {
	switch p := Priority(s); p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return p, nil
	}
	return "", shared.ErrInvalid
}

// Kind clasifica una columna dentro del flujo de trabajo y define cuándo corre el cycle time.
type Kind string

const (
	KindNormal Kind = "normal"
	KindActive Kind = "active" // al entrar empieza el cycle time
	KindDone   Kind = "done"   // al entrar termina el cycle time
)

func ParseKind(s string) (Kind, error) {
	switch k := Kind(s); k {
	case KindNormal, KindActive, KindDone:
		return k, nil
	}
	return "", shared.ErrInvalid
}

// WIPLimit es el límite de trabajo en progreso de una columna. El valor cero significa "sin límite".
type WIPLimit struct{ n int }

func NoWIPLimit() WIPLimit { return WIPLimit{} }

func NewWIPLimit(n int) (WIPLimit, error) {
	if n < 1 {
		return WIPLimit{}, shared.ErrInvalid
	}
	return WIPLimit{n: n}, nil
}

// Value devuelve el límite y si está definido.
func (l WIPLimit) Value() (int, bool) { return l.n, l.n > 0 }

// Allows indica si una columna con currentCount tarjetas puede aceptar una más.
func (l WIPLimit) Allows(currentCount int) bool {
	return l.n == 0 || currentCount < l.n
}
