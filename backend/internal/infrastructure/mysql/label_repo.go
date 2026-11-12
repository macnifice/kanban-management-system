package mysql

import (
	"context"

	"kanban/internal/domain/kanban"
)

type labelRepo struct{ q dbtx }

func (r *labelRepo) Add(ctx context.Context, l *kanban.Label) error {
	res, err := r.q.ExecContext(ctx, `INSERT INTO labels (board_id, name, color) VALUES (?,?,?)`, l.BoardID, l.Name, l.Color)
	if err != nil {
		return mapErr(err)
	}
	id, _ := res.LastInsertId()
	l.ID = int(id)
	return nil
}

func (r *labelRepo) Get(ctx context.Context, id int) (*kanban.Label, error) {
	var l kanban.Label
	err := r.q.QueryRowContext(ctx, `SELECT id, board_id, name, color FROM labels WHERE id=?`, id).
		Scan(&l.ID, &l.BoardID, &l.Name, &l.Color)
	if err != nil {
		return nil, mapErr(err)
	}
	return &l, nil
}

func (r *labelRepo) ListByBoard(ctx context.Context, boardID int) ([]*kanban.Label, error) {
	rows, err := r.q.QueryContext(ctx, `SELECT id, board_id, name, color FROM labels WHERE board_id=? ORDER BY id`, boardID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []*kanban.Label{}
	for rows.Next() {
		var l kanban.Label
		if err := rows.Scan(&l.ID, &l.BoardID, &l.Name, &l.Color); err != nil {
			return nil, err
		}
		out = append(out, &l)
	}
	return out, rows.Err()
}

func (r *labelRepo) Save(ctx context.Context, l *kanban.Label) error {
	_, err := r.q.ExecContext(ctx, `UPDATE labels SET name=?, color=? WHERE id=?`, l.Name, l.Color, l.ID)
	return mapErr(err)
}

func (r *labelRepo) Delete(ctx context.Context, id int) error {
	return affected(r.q.ExecContext(ctx, `DELETE FROM labels WHERE id=?`, id))
}
