package kanban

import (
	"time"

	"kanban/internal/domain/shared"
)

// Board es la raíz del agregado de tablero (proyecto).
type Board struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
}

func NewBoard(name, description string) (*Board, error) {
	n, err := shared.RequiredText(name, 120)
	if err != nil {
		return nil, err
	}
	return &Board{Name: n, Description: description}, nil
}

func (b *Board) Update(name, description string) error {
	n, err := shared.RequiredText(name, 120)
	if err != nil {
		return err
	}
	b.Name, b.Description = n, description
	return nil
}

// DefaultColumns es la fábrica de las columnas iniciales de un tablero nuevo.
func DefaultColumns(boardID int) []*Column {
	return []*Column{
		{BoardID: boardID, Title: "Por hacer", Kind: KindNormal},
		{BoardID: boardID, Title: "En progreso", Kind: KindActive},
		{BoardID: boardID, Title: "Hecho", Kind: KindDone},
	}
}
