package mysql

import (
	"context"
	"database/sql"
	"sort"

	"kanban/internal/domain/kanban"
)

type columnRepo struct{ q dbtx }

const columnCols = `id, board_id, title, position, wip_limit, kind`

type scanner interface{ Scan(dest ...any) error }

func scanColumn(s scanner) (*kanban.Column, error) {
	var c kanban.Column
	var wip sql.NullInt64
	var kind string
	if err := s.Scan(&c.ID, &c.BoardID, &c.Title, &c.Position, &wip, &kind); err != nil {
		return nil, mapErr(err)
	}
	c.Kind = kanban.Kind(kind)
	if wip.Valid {
		l, err := kanban.NewWIPLimit(int(wip.Int64))
		if err != nil {
			return nil, err
		}
		c.WIP = l
	}
	return &c, nil
}

func wipArg(l kanban.WIPLimit) any {
	if n, ok := l.Value(); ok {
		return n
	}
	return nil
}

func (r *columnRepo) Add(ctx context.Context, c *kanban.Column) error {
	res, err := r.q.ExecContext(ctx, `INSERT INTO columns_ (board_id, title, position, wip_limit, kind)
		SELECT ?, ?, COALESCE(MAX(position)+1, 0), ?, ? FROM columns_ WHERE board_id=?`,
		c.BoardID, c.Title, wipArg(c.WIP), string(c.Kind), c.BoardID)
	if err != nil {
		return mapErr(err)
	}
	id, _ := res.LastInsertId()
	c.ID = int(id)
	return mapErr(r.q.QueryRowContext(ctx, `SELECT position FROM columns_ WHERE id=?`, id).Scan(&c.Position))
}

func (r *columnRepo) Get(ctx context.Context, id int) (*kanban.Column, error) {
	return scanColumn(r.q.QueryRowContext(ctx, `SELECT `+columnCols+` FROM columns_ WHERE id=?`, id))
}

func (r *columnRepo) Lock(ctx context.Context, ids ...int) (map[int]*kanban.Column, error) {
	sorted := append([]int(nil), ids...)
	sort.Ints(sorted)
	out := map[int]*kanban.Column{}
	for _, id := range sorted {
		if _, done := out[id]; done {
			continue
		}
		c, err := scanColumn(r.q.QueryRowContext(ctx, `SELECT `+columnCols+` FROM columns_ WHERE id=? FOR UPDATE`, id))
		if err != nil {
			return nil, err
		}
		out[id] = c
	}
	return out, nil
}

func (r *columnRepo) Save(ctx context.Context, c *kanban.Column) error {
	_, err := r.q.ExecContext(ctx, `UPDATE columns_ SET title=?, wip_limit=?, kind=? WHERE id=?`,
		c.Title, wipArg(c.WIP), string(c.Kind), c.ID)
	return mapErr(err)
}

func (r *columnRepo) Delete(ctx context.Context, id int) error {
	return affected(r.q.ExecContext(ctx, `DELETE FROM columns_ WHERE id=?`, id))
}

func (r *columnRepo) Reorder(ctx context.Context, boardID int, ids []int) error {
	for i, id := range ids {
		if _, err := r.q.ExecContext(ctx, `UPDATE columns_ SET position=? WHERE id=? AND board_id=?`, i, id, boardID); err != nil {
			return mapErr(err)
		}
	}
	return nil
}
