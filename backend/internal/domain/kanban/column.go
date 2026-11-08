package kanban

import "kanban/internal/domain/shared"

// Column es la raíz del agregado de columna: etapa del flujo con su límite de WIP.
type Column struct {
	ID       int
	BoardID  int
	Title    string
	Position int
	WIP      WIPLimit
	Kind     Kind
}

func NewColumn(boardID int, title string) (*Column, error) {
	t, err := shared.RequiredText(title, 100)
	if err != nil {
		return nil, err
	}
	return &Column{BoardID: boardID, Title: t, Kind: KindNormal}, nil
}

func (c *Column) Rename(title string) error {
	t, err := shared.RequiredText(title, 100)
	if err != nil {
		return err
	}
	c.Title = t
	return nil
}

func (c *Column) SetWIPLimit(l WIPLimit) { c.WIP = l }
func (c *Column) SetKind(k Kind)         { c.Kind = k }

// CanAccept aplica la regla de WIP: la columna no puede recibir más tarjetas si ya alcanzó su límite.
func (c *Column) CanAccept(cardsInColumn int) error {
	if !c.WIP.Allows(cardsInColumn) {
		return ErrWIPLimitReached
	}
	return nil
}
