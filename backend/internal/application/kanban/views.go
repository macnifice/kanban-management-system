// Package kanban (aplicación) orquesta los casos de uso del contexto kanban.
// Los *View son modelos de lectura que forman el contrato publicado hacia la API.
package kanban

import (
	"time"

	domain "kanban/internal/domain/kanban"
)

type BoardSummary struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	CardCount   int       `json:"card_count"`
}

type LabelView struct {
	ID      int    `json:"id"`
	BoardID int    `json:"board_id"`
	Name    string `json:"name"`
	Color   string `json:"color"`
}

type CardView struct {
	ID          int       `json:"id"`
	ColumnID    int       `json:"column_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Position    int       `json:"position"`
	Priority    string    `json:"priority"`
	AssigneeID  *int      `json:"assignee_id"`
	LabelIDs    []int     `json:"label_ids"`
	CreatedAt   time.Time `json:"created_at"`
}

type ColumnView struct {
	ID       int        `json:"id"`
	BoardID  int        `json:"board_id"`
	Title    string     `json:"title"`
	Position int        `json:"position"`
	WIPLimit *int       `json:"wip_limit"`
	Kind     string     `json:"kind"`
	Cards    []CardView `json:"cards"`
}

type BoardView struct {
	BoardSummary
	Columns []ColumnView `json:"columns"`
	Labels  []LabelView  `json:"labels"`
}

type MovementView struct {
	ID         int64     `json:"id"`
	MovedAt    time.Time `json:"moved_at"`
	FromColumn *string   `json:"from_column"`
	ToColumn   string    `json:"to_column"`
	ToKind     string    `json:"to_kind"`
	User       *string   `json:"user"`
}

type ColumnMetric struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Kind     string `json:"kind"`
	Count    int    `json:"count"`
	WIPLimit *int   `json:"wip_limit"`
}

type CycleSampleView struct {
	CardID int     `json:"card_id"`
	Title  string  `json:"title"`
	Hours  float64 `json:"hours"`
}

type CycleTimeView struct {
	AvgHours    float64           `json:"avg_hours"`
	MedianHours float64           `json:"median_hours"`
	P85Hours    float64           `json:"p85_hours"`
	Count       int               `json:"count"`
	Recent      []CycleSampleView `json:"recent"`
}

type MetricsView struct {
	Columns      []ColumnMetric `json:"columns"`
	TotalsByKind map[string]int `json:"totals_by_kind"`
	TotalCards   int            `json:"total_cards"`
	CycleTime    CycleTimeView  `json:"cycle_time"`
}

func BoardSummaryOf(b *domain.Board, cardCount int) BoardSummary {
	return BoardSummary{ID: b.ID, Name: b.Name, Description: b.Description, CreatedAt: b.CreatedAt, CardCount: cardCount}
}

func LabelViewOf(l *domain.Label) LabelView {
	return LabelView{ID: l.ID, BoardID: l.BoardID, Name: l.Name, Color: l.Color}
}

func CardViewOf(c *domain.Card) CardView {
	ids := c.LabelIDs
	if ids == nil {
		ids = []int{}
	}
	return CardView{
		ID: c.ID, ColumnID: c.ColumnID, Title: c.Title, Description: c.Description, Position: c.Position,
		Priority: string(c.Priority), AssigneeID: c.AssigneeID, LabelIDs: ids, CreatedAt: c.CreatedAt,
	}
}

func ColumnViewOf(c *domain.Column) ColumnView {
	var wip *int
	if n, ok := c.WIP.Value(); ok {
		wip = &n
	}
	return ColumnView{
		ID: c.ID, BoardID: c.BoardID, Title: c.Title, Position: c.Position,
		WIPLimit: wip, Kind: string(c.Kind), Cards: []CardView{},
	}
}
