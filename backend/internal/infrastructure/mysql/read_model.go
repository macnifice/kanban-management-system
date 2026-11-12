package mysql

import (
	"context"
	"database/sql"
	"time"

	app "kanban/internal/application/kanban"
	"kanban/internal/domain/kanban"
)

// ReadModel implementa app.ReadModel con consultas SQL directas (lado query de CQRS).
type ReadModel struct{ db *sql.DB }

func NewReadModel(db *sql.DB) *ReadModel { return &ReadModel{db: db} }

func (r *ReadModel) boardExists(ctx context.Context, id int) error {
	var one int
	return mapErr(r.db.QueryRowContext(ctx, `SELECT 1 FROM boards WHERE id=?`, id).Scan(&one))
}

func (r *ReadModel) ListBoards(ctx context.Context) ([]app.BoardSummary, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT b.id, b.name, COALESCE(b.description,''), b.created_at,
		(SELECT COUNT(*) FROM cards c JOIN columns_ col ON col.id=c.column_id WHERE col.board_id=b.id)
		FROM boards b ORDER BY b.id`)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []app.BoardSummary{}
	for rows.Next() {
		var b app.BoardSummary
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.CreatedAt, &b.CardCount); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *ReadModel) Board(ctx context.Context, id int) (*app.BoardView, error) {
	v := &app.BoardView{Columns: []app.ColumnView{}, Labels: []app.LabelView{}}
	err := r.db.QueryRowContext(ctx, `SELECT id, name, COALESCE(description,''), created_at FROM boards WHERE id=?`, id).
		Scan(&v.ID, &v.Name, &v.Description, &v.CreatedAt)
	if err != nil {
		return nil, mapErr(err)
	}

	colIdx := map[int]int{}
	rows, err := r.db.QueryContext(ctx, `SELECT `+columnCols+` FROM columns_ WHERE board_id=? ORDER BY position, id`, id)
	if err != nil {
		return nil, mapErr(err)
	}
	for rows.Next() {
		c, err := scanColumn(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		colIdx[c.ID] = len(v.Columns)
		v.Columns = append(v.Columns, app.ColumnViewOf(c))
	}
	rows.Close()

	all := []app.CardView{}
	cardIdx := map[int]int{}
	crows, err := r.db.QueryContext(ctx, `SELECT c.id, c.column_id, c.title, COALESCE(c.description,''), c.position, c.priority, c.assignee_id, c.created_at
		FROM cards c JOIN columns_ col ON col.id=c.column_id WHERE col.board_id=? ORDER BY c.position, c.id`, id)
	if err != nil {
		return nil, mapErr(err)
	}
	for crows.Next() {
		c := app.CardView{LabelIDs: []int{}}
		var assignee sql.NullInt64
		if err := crows.Scan(&c.ID, &c.ColumnID, &c.Title, &c.Description, &c.Position, &c.Priority, &assignee, &c.CreatedAt); err != nil {
			crows.Close()
			return nil, err
		}
		if assignee.Valid {
			a := int(assignee.Int64)
			c.AssigneeID = &a
		}
		cardIdx[c.ID] = len(all)
		all = append(all, c)
	}
	crows.Close()

	lrows, err := r.db.QueryContext(ctx, `SELECT cl.card_id, cl.label_id FROM card_labels cl
		JOIN cards c ON c.id=cl.card_id JOIN columns_ col ON col.id=c.column_id WHERE col.board_id=? ORDER BY cl.label_id`, id)
	if err != nil {
		return nil, mapErr(err)
	}
	for lrows.Next() {
		var cid, lid int
		if err := lrows.Scan(&cid, &lid); err != nil {
			lrows.Close()
			return nil, err
		}
		if i, ok := cardIdx[cid]; ok {
			all[i].LabelIDs = append(all[i].LabelIDs, lid)
		}
	}
	lrows.Close()

	for _, c := range all {
		if i, ok := colIdx[c.ColumnID]; ok {
			v.Columns[i].Cards = append(v.Columns[i].Cards, c)
		}
	}

	labels, err := (&labelRepo{r.db}).ListByBoard(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, l := range labels {
		v.Labels = append(v.Labels, app.LabelViewOf(l))
	}
	return v, nil
}

func (r *ReadModel) CardHistory(ctx context.Context, cardID int) ([]app.MovementView, error) {
	var one int
	if err := r.db.QueryRowContext(ctx, `SELECT 1 FROM cards WHERE id=?`, cardID).Scan(&one); err != nil {
		return nil, mapErr(err)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT e.id, e.moved_at, fc.title, COALESCE(tc.title, '(columna eliminada)'), e.to_kind, u.name
		FROM card_events e
		LEFT JOIN columns_ fc ON fc.id=e.from_column_id
		LEFT JOIN columns_ tc ON tc.id=e.to_column_id
		LEFT JOIN users u ON u.id=e.user_id
		WHERE e.card_id=? ORDER BY e.moved_at, e.id`, cardID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []app.MovementView{}
	for rows.Next() {
		var e app.MovementView
		var from, user sql.NullString
		if err := rows.Scan(&e.ID, &e.MovedAt, &from, &e.ToColumn, &e.ToKind, &user); err != nil {
			return nil, err
		}
		if from.Valid {
			e.FromColumn = &from.String
		}
		if user.Valid {
			e.User = &user.String
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *ReadModel) ColumnCounts(ctx context.Context, boardID int) ([]app.ColumnMetric, error) {
	if err := r.boardExists(ctx, boardID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT col.id, col.title, col.kind, col.wip_limit, COUNT(c.id)
		FROM columns_ col LEFT JOIN cards c ON c.column_id=col.id
		WHERE col.board_id=? GROUP BY col.id ORDER BY col.position, col.id`, boardID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []app.ColumnMetric{}
	for rows.Next() {
		var m app.ColumnMetric
		var wip sql.NullInt64
		if err := rows.Scan(&m.ID, &m.Title, &m.Kind, &wip, &m.Count); err != nil {
			return nil, err
		}
		if wip.Valid {
			w := int(wip.Int64)
			m.WIPLimit = &w
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// CompletedCards devuelve, por cada tarjeta en una columna "done", su primera entrada a "active"
// y su última entrada a "done" según el historial de movimientos.
func (r *ReadModel) CompletedCards(ctx context.Context, boardID int) ([]kanban.CompletedCard, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT c.id, c.title,
		(SELECT MIN(e.moved_at) FROM card_events e WHERE e.card_id=c.id AND e.to_kind='active'),
		(SELECT MAX(e.moved_at) FROM card_events e WHERE e.card_id=c.id AND e.to_kind='done')
		FROM cards c JOIN columns_ col ON col.id=c.column_id
		WHERE col.board_id=? AND col.kind='done'`, boardID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []kanban.CompletedCard{}
	for rows.Next() {
		var cc kanban.CompletedCard
		var start, end sql.NullTime
		if err := rows.Scan(&cc.CardID, &cc.Title, &start, &end); err != nil {
			return nil, err
		}
		cc.StartedAt, cc.FinishedAt = nullTime(start), nullTime(end)
		out = append(out, cc)
	}
	return out, rows.Err()
}

func nullTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
