package kanban

import "kanban/internal/domain/shared"

type Label struct {
	ID      int
	BoardID int
	Name    string
	Color   string // #rrggbb
}

func NewLabel(boardID int, name, color string) (*Label, error) {
	l := &Label{BoardID: boardID}
	if err := l.Update(name, color); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Label) Update(name, color string) error {
	n, err := shared.RequiredText(name, 60)
	if err != nil {
		return err
	}
	if !validColor(color) {
		return shared.ErrInvalid
	}
	l.Name, l.Color = n, color
	return nil
}

func validColor(c string) bool {
	if len(c) != 7 || c[0] != '#' {
		return false
	}
	for _, ch := range c[1:] {
		if !(ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F') {
			return false
		}
	}
	return true
}
