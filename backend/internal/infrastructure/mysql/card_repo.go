package mysql

import (
	"context"
	"database/sql"

	"kanban/internal/domain/kanban"
)

type cardRepo struct{ q dbtx }

const cardSelect = `SELECT id, column_id, title, COALESCE(description,''), position, priority, assignee_id, created_at FROM cards WHERE id=?`

func (r *cardRepo) Add(ctx context.Context, c *kanban.Card) error {
	res, err := r.q.ExecContext(ctx, `INSERT INTO cards (column_id, title, description, position, priority, assignee_id)
		SELECT ?, ?, ?, COALESCE(MAX(position)+1, 0), ?, ? FROM cards WHERE column_id=?`,
		c.ColumnID, c.Title, c.Description, string(c.Priority), ptrArg(c.AssigneeID), c.ColumnID)
	if err != nil {
		return mapErr(err)
	}
	id, _ := res.LastInsertId()
	c.ID = int(id)
	if err := r.q.QueryRowContext(ctx, `SELECT position, created_at FROM cards WHERE id=?`, id).Scan(&c.Position, &c.CreatedAt); err != nil {
		return mapErr(err)
	}
	if err := r.setLabels(ctx, c); err != nil {
		return err
	}
	return r.persistMovements(ctx, c)
}

func (r *cardRepo) Get(ctx context.Context, id int) (*kanban.Card, error) {
	return r.load(ctx, id, false)
}

func (r *cardRepo) GetForUpdate(ctx context.Context, id int) (*kanban.Card, error) {
	return r.load(ctx, id, true)
}

func (r *cardRepo) load(ctx context.Context, id int, forUpdate bool) (*kanban.Card, error) {
	q := cardSelect
	if forUpdate {
		q += ` FOR UPDATE`
	}
	var c kanban.Card
	var prio string
	var assignee sql.NullInt64
	if err := r.q.QueryRowContext(ctx, q, id).
		Scan(&c.ID, &c.ColumnID, &c.Title, &c.Description, &c.Position, &prio, &assignee, &c.CreatedAt); err != nil {
		return nil, mapErr(err)
	}
	c.Priority = kanban.Priority(prio)
	if assignee.Valid {
		a := int(assignee.Int64)
		c.AssigneeID = &a
	}
	c.LabelIDs = []int{}
	rows, err := r.q.QueryContext(ctx, `SELECT label_id FROM card_labels WHERE card_id=? ORDER BY label_id`, id)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var l int
		if err := rows.Scan(&l); err != nil {
			return nil, err
		}
		c.LabelIDs = append(c.LabelIDs, l)
	}
	return &c, rows.Err()
}

func (r *cardRepo) Save(ctx context.Context, c *kanban.Card) error {
	if _, err := r.q.ExecContext(ctx, `UPDATE cards SET title=?, description=?, priority=?, assignee_id=? WHERE id=?`,
		c.Title, c.Description, string(c.Priority), ptrArg(c.AssigneeID), c.ID); err != nil {
		return mapErr(err)
	}
	return r.setLabels(ctx, c)
}

func (r *cardRepo) Delete(ctx context.Context, id int) error {
	return affected(r.q.ExecContext(ctx, `DELETE FROM cards WHERE id=?`, id))
}

func (r *cardRepo) CountInColumn(ctx context.Context, columnID, excludeID int) (int, error) {
	var n int
	err := r.q.QueryRowContext(ctx, `SELECT COUNT(*) FROM cards WHERE column_id=? AND id<>?`, columnID, excludeID).Scan(&n)
	return n, mapErr(err)
}

// Relocate guarda el cambio de columna ya aplicado en memoria: inserta la tarjeta en la posición
// pedida de la columna destino y recompacta las posiciones de la columna de origen.
func (r *cardRepo) Relocate(ctx context.Context, c *kanban.Card, position int) error {
	var oldCol int
	if err := r.q.QueryRowContext(ctx, `SELECT column_id FROM cards WHERE id=?`, c.ID).Scan(&oldCol); err != nil {
		return mapErr(err)
	}
	ids, err := r.cardIDs(ctx, c.ColumnID, c.ID)
	if err != nil {
		return err
	}
	if position < 0 {
		position = 0
	}
	if position > len(ids) {
		position = len(ids)
	}
	ids = append(ids[:position], append([]int{c.ID}, ids[position:]...)...)
	for i, id := range ids {
		if _, err := r.q.ExecContext(ctx, `UPDATE cards SET column_id=?, position=? WHERE id=?`, c.ColumnID, i, id); err != nil {
			return mapErr(err)
		}
	}
	c.Position = position
	if oldCol != c.ColumnID {
		rest, err := r.cardIDs(ctx, oldCol, c.ID)
		if err != nil {
			return err
		}
		for i, id := range rest {
			if _, err := r.q.ExecContext(ctx, `UPDATE cards SET position=? WHERE id=?`, i, id); err != nil {
				return mapErr(err)
			}
		}
	}
	return r.persistMovements(ctx, c)
}

func (r *cardRepo) cardIDs(ctx context.Context, columnID, exclude int) ([]int, error) {
	rows, err := r.q.QueryContext(ctx, `SELECT id FROM cards WHERE column_id=? AND id<>? ORDER BY position, id`, columnID, exclude)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	ids := []int{}
	for rows.Next() {
		var i int
		if err := rows.Scan(&i); err != nil {
			return nil, err
		}
		ids = append(ids, i)
	}
	return ids, rows.Err()
}

func (r *cardRepo) setLabels(ctx context.Context, c *kanban.Card) error {
	if _, err := r.q.ExecContext(ctx, `DELETE FROM card_labels WHERE card_id=?`, c.ID); err != nil {
		return mapErr(err)
	}
	for _, l := range c.LabelIDs {
		if _, err := r.q.ExecContext(ctx, `INSERT IGNORE INTO card_labels (card_id, label_id) VALUES (?, ?)`, c.ID, l); err != nil {
			return mapErr(err)
		}
	}
	return nil
}

func (r *cardRepo) persistMovements(ctx context.Context, c *kanban.Card) error {
	for _, m := range c.PullMovements() {
		if _, err := r.q.ExecContext(ctx,
			`INSERT INTO card_events (card_id, from_column_id, to_column_id, to_kind, user_id) VALUES (?,?,?,?,?)`,
			c.ID, ptrArg(m.FromColumnID), m.ToColumnID, string(m.ToKind), nilIfZero(m.UserID)); err != nil {
			return mapErr(err)
		}
	}
	return nil
}
