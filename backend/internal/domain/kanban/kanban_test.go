package kanban

import (
	"errors"
	"testing"
	"time"

	"kanban/internal/domain/shared"
)

func col(id, board int, kind Kind, limit int) *Column {
	c := &Column{ID: id, BoardID: board, Title: "c", Kind: kind}
	if limit > 0 {
		c.WIP, _ = NewWIPLimit(limit)
	}
	return c
}

func TestWIPLimit(t *testing.T) {
	if !NoWIPLimit().Allows(1000) {
		t.Fatal("sin límite siempre permite")
	}
	l, _ := NewWIPLimit(2)
	if !l.Allows(1) || l.Allows(2) || l.Allows(3) {
		t.Fatal("con límite 2 solo permite hasta 1 tarjeta existente")
	}
	if _, err := NewWIPLimit(0); !errors.Is(err, shared.ErrInvalid) {
		t.Fatal("el límite debe ser >= 1")
	}
}

func TestMoveCardRespectsWIP(t *testing.T) {
	from, to := col(1, 1, KindNormal, 0), col(2, 1, KindActive, 1)
	card := &Card{ID: 10, ColumnID: 1}

	if err := MoveCard(card, from, to, 1, 7); !errors.Is(err, ErrWIPLimitReached) {
		t.Fatalf("esperaba ErrWIPLimitReached, obtuve %v", err)
	}
	if card.ColumnID != 1 || len(card.PullMovements()) != 0 {
		t.Fatal("un movimiento rechazado no debe cambiar la tarjeta ni registrar eventos")
	}

	if err := MoveCard(card, from, to, 0, 7); err != nil {
		t.Fatal(err)
	}
	ms := card.PullMovements()
	if card.ColumnID != 2 || len(ms) != 1 || ms[0].ToKind != KindActive || ms[0].UserID != 7 || *ms[0].FromColumnID != 1 {
		t.Fatalf("movimiento mal registrado: %+v", ms)
	}
	if len(card.PullMovements()) != 0 {
		t.Fatal("PullMovements debe vaciar los eventos")
	}
}

func TestReorderInsideFullColumnIsAllowed(t *testing.T) {
	full := col(1, 1, KindActive, 1)
	card := &Card{ID: 1, ColumnID: 1}
	if err := MoveCard(card, full, full, 1, 1); err != nil {
		t.Fatalf("reordenar dentro de la misma columna debe permitirse: %v", err)
	}
	if len(card.PullMovements()) != 0 {
		t.Fatal("no hay evento si no cambia de columna")
	}
}

func TestMoveAcrossBoardsIsInvalid(t *testing.T) {
	card := &Card{ID: 1, ColumnID: 1}
	if err := MoveCard(card, col(1, 1, KindNormal, 0), col(2, 2, KindNormal, 0), 0, 1); !errors.Is(err, shared.ErrInvalid) {
		t.Fatalf("esperaba ErrInvalid, obtuve %v", err)
	}
}

func TestCreateCard(t *testing.T) {
	full := col(1, 1, KindNormal, 1)
	if _, err := CreateCard(full, 1, Draft{Title: "x"}, 1); !errors.Is(err, ErrWIPLimitReached) {
		t.Fatalf("crear en columna llena debe fallar, obtuve %v", err)
	}
	c, err := CreateCard(col(5, 1, KindActive, 0), 0, Draft{Title: "  hola  "}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "hola" || c.Priority != PriorityMedium || c.ColumnID != 5 {
		t.Fatalf("tarjeta inesperada: %+v", c)
	}
	ms := c.PullMovements()
	if len(ms) != 1 || ms[0].FromColumnID != nil || ms[0].ToKind != KindActive {
		t.Fatalf("la creación debe registrar un movimiento inicial: %+v", ms)
	}
	if _, err := CreateCard(col(1, 1, KindNormal, 0), 0, Draft{Title: "   "}, 1); !errors.Is(err, shared.ErrInvalid) {
		t.Fatal("el título vacío es inválido")
	}
}

func TestSetLabelsIgnoresForeignLabels(t *testing.T) {
	c := &Card{}
	c.SetLabels([]int{1, 2, 2, 9}, []*Label{{ID: 1}, {ID: 2}})
	if len(c.LabelIDs) != 2 || c.LabelIDs[0] != 1 || c.LabelIDs[1] != 2 {
		t.Fatalf("etiquetas: %v", c.LabelIDs)
	}
}

func TestLabelColorValidation(t *testing.T) {
	if _, err := NewLabel(1, "bug", "#ff0000"); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"red", "#ff00", "ff0000", "#gg0000"} {
		if _, err := NewLabel(1, "bug", bad); !errors.Is(err, shared.ErrInvalid) {
			t.Fatalf("%q debería ser inválido", bad)
		}
	}
}

func TestComputeCycleTime(t *testing.T) {
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	at := func(h int) *time.Time { x := base.Add(time.Duration(h) * time.Hour); return &x }
	st := ComputeCycleTime([]CompletedCard{
		{CardID: 1, Title: "a", StartedAt: at(0), FinishedAt: at(2)},  // 2h
		{CardID: 2, Title: "b", StartedAt: at(0), FinishedAt: at(10)}, // 10h
		{CardID: 3, Title: "sin inicio", StartedAt: nil, FinishedAt: at(5)},
		{CardID: 4, Title: "c", StartedAt: at(1), FinishedAt: at(7)}, // 6h, la más reciente es la 2
		{CardID: 5, Title: "invertida", StartedAt: at(9), FinishedAt: at(3)},
	}, 2)
	if st.Count != 3 {
		t.Fatalf("count = %d, esperaba 3", st.Count)
	}
	if st.Average != 6*time.Hour || st.Median != 6*time.Hour {
		t.Fatalf("avg=%v median=%v, esperaba 6h/6h", st.Average, st.Median)
	}
	if len(st.Recent) != 2 || st.Recent[0].CardID != 2 || st.Recent[1].CardID != 4 {
		t.Fatalf("recientes: %+v", st.Recent)
	}
	if empty := ComputeCycleTime(nil, 5); empty.Count != 0 || empty.Recent == nil {
		t.Fatal("sin datos debe devolver ceros y slice vacío")
	}
}

func TestMedianEvenCount(t *testing.T) {
	base := time.Now()
	f := func(h int) *time.Time { x := base.Add(time.Duration(h) * time.Hour); return &x }
	st := ComputeCycleTime([]CompletedCard{
		{CardID: 1, StartedAt: &base, FinishedAt: f(2)},
		{CardID: 2, StartedAt: &base, FinishedAt: f(4)},
	}, 5)
	if st.Median != 3*time.Hour {
		t.Fatalf("mediana = %v", st.Median)
	}
}
