package mysql

import (
	"context"

	"kanban/internal/domain/kanban"
)

type boardRepo struct{ q dbtx }

func (r *boardRepo) Add(ctx context.Context, b *kanban.Board) error {
	res, err := r.q.ExecContext(ctx, `INSERT INTO boards (name, description) VALUES (?, ?)`, b.Name, b.Description)
	if err != nil {
		return mapErr(err)
	}
	id, _ := res.LastInsertId()
	b.ID = int(id)
	return mapErr(r.q.QueryRowContext(ctx, `SELECT created_at FROM boards WHERE id=?`, id).Scan(&b.CreatedAt))
}

func (r *boardRepo) Get(ctx context.Context, id int) (*kanban.Board, error) {
	var b kanban.Board
	err := r.q.QueryRowContext(ctx, `SELECT id, name, COALESCE(description,''), created_at FROM boards WHERE id=?`, id).
		Scan(&b.ID, &b.Name, &b.Description, &b.CreatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &b, nil
}

func (r *boardRepo) Save(ctx context.Context, b *kanban.Board) error {
	_, err := r.q.ExecContext(ctx, `UPDATE boards SET name=?, description=? WHERE id=?`, b.Name, b.Description, b.ID)
	return mapErr(err)
}

func (r *boardRepo) Delete(ctx context.Context, id int) error {
	return affected(r.q.ExecContext(ctx, `DELETE FROM boards WHERE id=?`, id))
}
