package kanban

import "context"

// Puertos de persistencia del dominio. Los implementa la infraestructura.
// Todas las operaciones de un UnitOfWork.Do comparten la misma transacción.

type BoardRepository interface {
	Add(ctx context.Context, b *Board) error // asigna ID y CreatedAt
	Get(ctx context.Context, id int) (*Board, error)
	Save(ctx context.Context, b *Board) error
	Delete(ctx context.Context, id int) error
}

type ColumnRepository interface {
	Add(ctx context.Context, c *Column) error // la coloca al final del tablero; asigna ID y Position
	Get(ctx context.Context, id int) (*Column, error)
	// Lock bloquea las columnas (siempre en orden de ID, para evitar deadlocks) y las devuelve por ID.
	Lock(ctx context.Context, ids ...int) (map[int]*Column, error)
	Save(ctx context.Context, c *Column) error
	Delete(ctx context.Context, id int) error
	Reorder(ctx context.Context, boardID int, ids []int) error
}

type CardRepository interface {
	// Add coloca la tarjeta al final de su columna, asigna ID/Position/CreatedAt y persiste sus movimientos.
	Add(ctx context.Context, c *Card) error
	Get(ctx context.Context, id int) (*Card, error)
	GetForUpdate(ctx context.Context, id int) (*Card, error)
	Save(ctx context.Context, c *Card) error
	Delete(ctx context.Context, id int) error
	// CountInColumn cuenta las tarjetas de una columna sin contar excludeID (0 = no excluir).
	CountInColumn(ctx context.Context, columnID, excludeID int) (int, error)
	// Relocate persiste el cambio de columna de la tarjeta en la posición dada, recompacta las
	// posiciones de origen y destino y persiste los movimientos pendientes.
	Relocate(ctx context.Context, c *Card, position int) error
}

type LabelRepository interface {
	Add(ctx context.Context, l *Label) error
	Get(ctx context.Context, id int) (*Label, error)
	ListByBoard(ctx context.Context, boardID int) ([]*Label, error)
	Save(ctx context.Context, l *Label) error
	Delete(ctx context.Context, id int) error
}

// Repositories son los repositorios ligados a una transacción.
type Repositories interface {
	Boards() BoardRepository
	Columns() ColumnRepository
	Cards() CardRepository
	Labels() LabelRepository
}

// UnitOfWork ejecuta fn dentro de una transacción; se confirma si fn devuelve nil.
// La implementación puede reejecutar fn ante deadlocks o ErrConcurrentChange.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context, r Repositories) error) error
}
