package kanban

import (
	"time"

	"kanban/internal/domain/shared"
)

// Movement es el evento de dominio "la tarjeta entró a una columna". Es la base del historial y del cycle time.
// Su instante lo fija la persistencia al guardarlo.
type Movement struct {
	FromColumnID *int // nil cuando la tarjeta se crea
	ToColumnID   int
	ToKind       Kind // instantánea: sobrevive a renombrar o borrar la columna
	UserID       int  // 0 = desconocido
}

// Card es la raíz del agregado de tarjeta.
type Card struct {
	ID          int
	ColumnID    int
	Title       string
	Description string
	Position    int
	Priority    Priority
	AssigneeID  *int
	LabelIDs    []int
	CreatedAt   time.Time

	movements []Movement
}

// Draft agrupa los datos para crear una tarjeta.
type Draft struct {
	Title       string
	Description string
	Priority    Priority
	AssigneeID  *int
	LabelIDs    []int
}

func newCard(col *Column, d Draft, by int) (*Card, error) {
	title, err := shared.RequiredText(d.Title, 200)
	if err != nil {
		return nil, err
	}
	if d.Priority == "" {
		d.Priority = PriorityMedium
	}
	if _, err := ParsePriority(string(d.Priority)); err != nil {
		return nil, err
	}
	ids := d.LabelIDs
	if ids == nil {
		ids = []int{}
	}
	c := &Card{
		ColumnID: col.ID, Title: title, Description: d.Description,
		Priority: d.Priority, AssigneeID: d.AssigneeID, LabelIDs: ids,
	}
	c.movements = append(c.movements, Movement{ToColumnID: col.ID, ToKind: col.Kind, UserID: by})
	return c, nil
}

func (c *Card) Rename(title string) error {
	t, err := shared.RequiredText(title, 200)
	if err != nil {
		return err
	}
	c.Title = t
	return nil
}

func (c *Card) Describe(text string)   { c.Description = text }
func (c *Card) SetPriority(p Priority) { c.Priority = p }
func (c *Card) Assign(userID *int)     { c.AssigneeID = userID }

// SetLabels conserva solo las etiquetas que pertenecen al tablero (allowed).
func (c *Card) SetLabels(ids []int, allowed []*Label) {
	ok := map[int]bool{}
	for _, l := range allowed {
		ok[l.ID] = true
	}
	out := []int{}
	seen := map[int]bool{}
	for _, id := range ids {
		if ok[id] && !seen[id] {
			out = append(out, id)
			seen[id] = true
		}
	}
	c.LabelIDs = out
}

func (c *Card) moveTo(col *Column, by int) {
	from := c.ColumnID
	c.ColumnID = col.ID
	c.movements = append(c.movements, Movement{FromColumnID: &from, ToColumnID: col.ID, ToKind: col.Kind, UserID: by})
}

// PullMovements entrega y vacía los eventos pendientes de persistir.
func (c *Card) PullMovements() []Movement {
	m := c.movements
	c.movements = nil
	return m
}

// CreateCard es el servicio de dominio que coloca una tarjeta nueva en una columna respetando el WIP.
// cardsInColumn es el número actual de tarjetas de la columna.
func CreateCard(col *Column, cardsInColumn int, d Draft, by int) (*Card, error) {
	if err := col.CanAccept(cardsInColumn); err != nil {
		return nil, err
	}
	return newCard(col, d, by)
}

// MoveCard es el servicio de dominio que mueve una tarjeta entre columnas del mismo tablero.
// cardsInTarget cuenta las tarjetas de destino sin incluir la que se mueve. Reordenar dentro de la
// misma columna siempre se permite.
func MoveCard(card *Card, from, to *Column, cardsInTarget int, by int) error {
	if card.ColumnID != from.ID || from.BoardID != to.BoardID {
		return shared.ErrInvalid
	}
	if from.ID == to.ID {
		return nil
	}
	if err := to.CanAccept(cardsInTarget); err != nil {
		return err
	}
	card.moveTo(to, by)
	return nil
}
