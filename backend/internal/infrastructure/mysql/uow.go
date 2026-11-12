package mysql

import (
	"context"
	"database/sql"

	"kanban/internal/domain/kanban"
)

// UnitOfWork implementa kanban.UnitOfWork con una transacción por llamada a Do,
// reintentada ante deadlocks o cambios concurrentes.
type UnitOfWork struct{ db *sql.DB }

func NewUnitOfWork(db *sql.DB) *UnitOfWork { return &UnitOfWork{db: db} }

func (u *UnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, r kanban.Repositories) error) error {
	return retry(func() error {
		// READ COMMITTED: cada sentencia ve lo último confirmado. Con REPEATABLE READ, una lectura previa al
		// bloqueo de la columna fijaría una instantánea vieja y el conteo de WIP podría ser incorrecto.
		// La consistencia de la regla de WIP la da el bloqueo explícito de las columnas (FOR UPDATE).
		tx, err := u.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if err := fn(ctx, &repositories{q: tx}); err != nil {
			return err
		}
		return tx.Commit()
	})
}

type repositories struct{ q dbtx }

func (r *repositories) Boards() kanban.BoardRepository   { return &boardRepo{r.q} }
func (r *repositories) Columns() kanban.ColumnRepository { return &columnRepo{r.q} }
func (r *repositories) Cards() kanban.CardRepository     { return &cardRepo{r.q} }
func (r *repositories) Labels() kanban.LabelRepository   { return &labelRepo{r.q} }
