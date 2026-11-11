package kanban

import (
	"context"

	domain "kanban/internal/domain/kanban"
	"kanban/internal/domain/shared"
)

// Nullable distingue "campo ausente" de "campo presente con valor null" en parches parciales.
type Nullable[T any] struct {
	Present bool
	Value   *T // nil = null explícito
}

type ColumnPatch struct {
	Title    *string
	WIPLimit Nullable[int]
	Kind     *string
}

// BoardService reúne los casos de uso de tableros y columnas.
type BoardService struct{ uow domain.UnitOfWork }

func NewBoardService(uow domain.UnitOfWork) *BoardService { return &BoardService{uow: uow} }

func (s *BoardService) CreateBoard(ctx context.Context, name, description string) (*domain.Board, error) {
	b, err := domain.NewBoard(name, description)
	if err != nil {
		return nil, err
	}
	err = s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		if err := r.Boards().Add(ctx, b); err != nil {
			return err
		}
		for _, c := range domain.DefaultColumns(b.ID) {
			if err := r.Columns().Add(ctx, c); err != nil {
				return err
			}
		}
		return nil
	})
	return b, err
}

func (s *BoardService) UpdateBoard(ctx context.Context, id int, name, description string) error {
	return s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		b, err := r.Boards().Get(ctx, id)
		if err != nil {
			return err
		}
		if err := b.Update(name, description); err != nil {
			return err
		}
		return r.Boards().Save(ctx, b)
	})
}

func (s *BoardService) DeleteBoard(ctx context.Context, id int) error {
	return s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		return r.Boards().Delete(ctx, id)
	})
}

func (s *BoardService) AddColumn(ctx context.Context, boardID int, title string) (*domain.Column, error) {
	var col *domain.Column
	err := s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		if _, err := r.Boards().Get(ctx, boardID); err != nil {
			return err
		}
		c, err := domain.NewColumn(boardID, title)
		if err != nil {
			return err
		}
		if err := r.Columns().Add(ctx, c); err != nil {
			return err
		}
		col = c
		return nil
	})
	return col, err
}

func (s *BoardService) UpdateColumn(ctx context.Context, id int, p ColumnPatch) error {
	if p.Title == nil && !p.WIPLimit.Present && p.Kind == nil {
		return shared.ErrInvalid
	}
	return s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		col, err := r.Columns().Get(ctx, id)
		if err != nil {
			return err
		}
		if p.Title != nil {
			if err := col.Rename(*p.Title); err != nil {
				return err
			}
		}
		if p.WIPLimit.Present {
			if p.WIPLimit.Value == nil {
				col.SetWIPLimit(domain.NoWIPLimit())
			} else {
				l, err := domain.NewWIPLimit(*p.WIPLimit.Value)
				if err != nil {
					return err
				}
				col.SetWIPLimit(l)
			}
		}
		if p.Kind != nil {
			k, err := domain.ParseKind(*p.Kind)
			if err != nil {
				return err
			}
			col.SetKind(k)
		}
		return r.Columns().Save(ctx, col)
	})
}

func (s *BoardService) DeleteColumn(ctx context.Context, id int) error {
	return s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		return r.Columns().Delete(ctx, id)
	})
}

func (s *BoardService) ReorderColumns(ctx context.Context, boardID int, ids []int) error {
	return s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		return r.Columns().Reorder(ctx, boardID, ids)
	})
}

// LabelService reúne los casos de uso de etiquetas.
type LabelService struct{ uow domain.UnitOfWork }

func NewLabelService(uow domain.UnitOfWork) *LabelService { return &LabelService{uow: uow} }

func (s *LabelService) Create(ctx context.Context, boardID int, name, color string) (*domain.Label, error) {
	var out *domain.Label
	err := s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		if _, err := r.Boards().Get(ctx, boardID); err != nil {
			return err
		}
		l, err := domain.NewLabel(boardID, name, color)
		if err != nil {
			return err
		}
		if err := r.Labels().Add(ctx, l); err != nil {
			return err
		}
		out = l
		return nil
	})
	return out, err
}

func (s *LabelService) Update(ctx context.Context, id int, name, color string) error {
	return s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		l, err := r.Labels().Get(ctx, id)
		if err != nil {
			return err
		}
		if err := l.Update(name, color); err != nil {
			return err
		}
		return r.Labels().Save(ctx, l)
	})
}

func (s *LabelService) Delete(ctx context.Context, id int) error {
	return s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		return r.Labels().Delete(ctx, id)
	})
}
