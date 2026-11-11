package kanban

import (
	"context"
	"time"

	domain "kanban/internal/domain/kanban"
)

// ReadModel es el puerto de lectura (lado "query"): consultas optimizadas que no pasan por los agregados.
type ReadModel interface {
	ListBoards(ctx context.Context) ([]BoardSummary, error)
	Board(ctx context.Context, id int) (*BoardView, error)
	CardHistory(ctx context.Context, cardID int) ([]MovementView, error)
	// ColumnCounts devuelve el número de tarjetas por columna; domain/shared.ErrNotFound si el tablero no existe.
	ColumnCounts(ctx context.Context, boardID int) ([]ColumnMetric, error)
	CompletedCards(ctx context.Context, boardID int) ([]domain.CompletedCard, error)
}

type QueryService struct{ rm ReadModel }

func NewQueryService(rm ReadModel) *QueryService { return &QueryService{rm: rm} }

func (q *QueryService) ListBoards(ctx context.Context) ([]BoardSummary, error) {
	return q.rm.ListBoards(ctx)
}

func (q *QueryService) Board(ctx context.Context, id int) (*BoardView, error) {
	return q.rm.Board(ctx, id)
}

func (q *QueryService) CardHistory(ctx context.Context, cardID int) ([]MovementView, error) {
	return q.rm.CardHistory(ctx, cardID)
}

// Metrics arma el dashboard: tarjetas por columna, WIP vs. límite, totales por estado y cycle time.
func (q *QueryService) Metrics(ctx context.Context, boardID int) (*MetricsView, error) {
	cols, err := q.rm.ColumnCounts(ctx, boardID)
	if err != nil {
		return nil, err
	}
	m := &MetricsView{Columns: cols, TotalsByKind: map[string]int{"normal": 0, "active": 0, "done": 0}}
	for _, c := range cols {
		m.TotalsByKind[c.Kind] += c.Count
		m.TotalCards += c.Count
	}
	done, err := q.rm.CompletedCards(ctx, boardID)
	if err != nil {
		return nil, err
	}
	st := domain.ComputeCycleTime(done, 10)
	m.CycleTime = CycleTimeView{
		AvgHours: hours(st.Average), MedianHours: hours(st.Median), Count: st.Count, Recent: []CycleSampleView{},
	}
	for _, s := range st.Recent {
		m.CycleTime.Recent = append(m.CycleTime.Recent, CycleSampleView{CardID: s.CardID, Title: s.Title, Hours: hours(s.Duration)})
	}
	return m, nil
}

func hours(d time.Duration) float64 { return d.Hours() }
