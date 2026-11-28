package kanban

import (
	"math"
	"sort"
	"time"
)

// CompletedCard son los datos mínimos de una tarjeta terminada para medir su cycle time.
// StartedAt es la primera entrada a una columna "active"; FinishedAt la última entrada a una "done".
type CompletedCard struct {
	CardID     int
	Title      string
	StartedAt  *time.Time
	FinishedAt *time.Time
}

type CycleSample struct {
	CardID   int
	Title    string
	Duration time.Duration
}

type CycleTimeStats struct {
	Count   int
	Average time.Duration
	Median  time.Duration
	P85     time.Duration // 85 % de las tarjetas se terminó en este tiempo o menos
	Recent  []CycleSample // las más recientemente terminadas primero
}

// ComputeCycleTime calcula las estadísticas ignorando tarjetas que nunca estuvieron "en curso".
func ComputeCycleTime(cards []CompletedCard, recentLimit int) CycleTimeStats {
	type item struct {
		CycleSample
		finished time.Time
	}
	items := []item{}
	for _, c := range cards {
		if c.StartedAt == nil || c.FinishedAt == nil || c.FinishedAt.Before(*c.StartedAt) {
			continue
		}
		items = append(items, item{CycleSample{c.CardID, c.Title, c.FinishedAt.Sub(*c.StartedAt)}, *c.FinishedAt})
	}
	st := CycleTimeStats{Count: len(items), Recent: []CycleSample{}}
	if len(items) == 0 {
		return st
	}
	ds := make([]time.Duration, len(items))
	var sum time.Duration
	for i, it := range items {
		ds[i] = it.Duration
		sum += it.Duration
	}
	sort.Slice(ds, func(i, j int) bool { return ds[i] < ds[j] })
	n := len(ds)
	if n%2 == 1 {
		st.Median = ds[n/2]
	} else {
		st.Median = (ds[n/2-1] + ds[n/2]) / 2
	}
	st.Average = sum / time.Duration(n)
	st.P85 = percentile(ds, 85)
	sort.Slice(items, func(i, j int) bool { return items[i].finished.After(items[j].finished) })
	for i := 0; i < len(items) && i < recentLimit; i++ {
		st.Recent = append(st.Recent, items[i].CycleSample)
	}
	return st
}

// percentile devuelve el percentil p (0-100) de duraciones ya ordenadas, con el
// método del rango más cercano: la duración en la posición ceil(p/100 * n).
// Una lista vacía da 0.
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	rank := int(math.Ceil(p / 100 * float64(len(sorted))))
	return sorted[rank-1]
}
