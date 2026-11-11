package kanban

import (
	"context"

	domain "kanban/internal/domain/kanban"
)

type CreateCardCommand struct {
	ColumnID    int
	Title       string
	Description string
	Priority    string // vacío = media
	AssigneeID  int    // 0 = sin asignar
	LabelIDs    []int
}

type CardPatch struct {
	Title       *string
	Description *string
	Priority    *string
	Assignee    Nullable[int]
	LabelIDs    *[]int
}

// CardService reúne los casos de uso de tarjetas.
type CardService struct{ uow domain.UnitOfWork }

func NewCardService(uow domain.UnitOfWork) *CardService { return &CardService{uow: uow} }

func (s *CardService) Create(ctx context.Context, cmd CreateCardCommand, userID int) (*domain.Card, error) {
	prio := domain.PriorityMedium
	if cmd.Priority != "" {
		p, err := domain.ParsePriority(cmd.Priority)
		if err != nil {
			return nil, err
		}
		prio = p
	}
	var card *domain.Card
	err := s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		cols, err := r.Columns().Lock(ctx, cmd.ColumnID) // serializa contra otros movimientos hacia la columna
		if err != nil {
			return err
		}
		col := cols[cmd.ColumnID]
		n, err := r.Cards().CountInColumn(ctx, col.ID, 0)
		if err != nil {
			return err
		}
		labels, err := r.Labels().ListByBoard(ctx, col.BoardID)
		if err != nil {
			return err
		}
		draft := domain.Draft{Title: cmd.Title, Description: cmd.Description, Priority: prio}
		if cmd.AssigneeID > 0 {
			a := cmd.AssigneeID
			draft.AssigneeID = &a
		}
		c, err := domain.CreateCard(col, n, draft, userID)
		if err != nil {
			return err
		}
		c.SetLabels(cmd.LabelIDs, labels)
		if err := r.Cards().Add(ctx, c); err != nil {
			return err
		}
		card = c
		return nil
	})
	return card, err
}

func (s *CardService) Update(ctx context.Context, id int, p CardPatch) (*domain.Card, error) {
	var card *domain.Card
	err := s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		c, err := r.Cards().GetForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if p.Title != nil {
			if err := c.Rename(*p.Title); err != nil {
				return err
			}
		}
		if p.Description != nil {
			c.Describe(*p.Description)
		}
		if p.Priority != nil {
			pr, err := domain.ParsePriority(*p.Priority)
			if err != nil {
				return err
			}
			c.SetPriority(pr)
		}
		if p.Assignee.Present {
			c.Assign(p.Assignee.Value)
		}
		if p.LabelIDs != nil {
			col, err := r.Columns().Get(ctx, c.ColumnID)
			if err != nil {
				return err
			}
			labels, err := r.Labels().ListByBoard(ctx, col.BoardID)
			if err != nil {
				return err
			}
			c.SetLabels(*p.LabelIDs, labels)
		}
		if err := r.Cards().Save(ctx, c); err != nil {
			return err
		}
		card = c
		return nil
	})
	return card, err
}

func (s *CardService) Delete(ctx context.Context, id int) error {
	return s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		return r.Cards().Delete(ctx, id)
	})
}

// Move mueve una tarjeta aplicando la regla de WIP.
// Orden de bloqueo: columnas (por ID) primero y la tarjeta después, para evitar deadlocks.
func (s *CardService) Move(ctx context.Context, id, toColumnID, position, userID int) error {
	return s.uow.Do(ctx, func(ctx context.Context, r domain.Repositories) error {
		peek, err := r.Cards().Get(ctx, id) // lectura sin bloqueo para conocer la columna origen
		if err != nil {
			return err
		}
		cols, err := r.Columns().Lock(ctx, peek.ColumnID, toColumnID)
		if err != nil {
			return err
		}
		card, err := r.Cards().GetForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if card.ColumnID != peek.ColumnID {
			return domain.ErrConcurrentChange // la tarjeta se movió mientras bloqueábamos
		}
		from, to := cols[card.ColumnID], cols[toColumnID]
		n, err := r.Cards().CountInColumn(ctx, to.ID, card.ID)
		if err != nil {
			return err
		}
		if err := domain.MoveCard(card, from, to, n, userID); err != nil {
			return err
		}
		return r.Cards().Relocate(ctx, card, position)
	})
}
