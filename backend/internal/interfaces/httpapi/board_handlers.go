package httpapi

import (
	"net/http"

	appkb "kanban/internal/application/kanban"
	"kanban/internal/domain/shared"
)

type boardReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Server) listBoards(w http.ResponseWriter, r *http.Request) {
	bs, err := s.Queries.ListBoards(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, 200, bs)
}

func (s *Server) createBoard(w http.ResponseWriter, r *http.Request) {
	var req boardReq
	if !decode(r, &req) {
		fail(w, 400, "petición inválida")
		return
	}
	b, err := s.Boards.CreateBoard(r.Context(), req.Name, req.Description)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, 201, appkb.BoardSummaryOf(b, 0))
}

func (s *Server) getBoard(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	if !valid {
		fail(w, 400, "id inválido")
		return
	}
	b, err := s.Queries.Board(r.Context(), id)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, 200, b)
}

func (s *Server) updateBoard(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	var req boardReq
	if !valid || !decode(r, &req) {
		fail(w, 400, "petición inválida")
		return
	}
	if err := s.Boards.UpdateBoard(r.Context(), id, req.Name, req.Description); err != nil {
		failErr(w, err)
		return
	}
	ok(w)
}

func (s *Server) deleteBoard(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	if !valid {
		fail(w, 400, "id inválido")
		return
	}
	if err := s.Boards.DeleteBoard(r.Context(), id); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	if !valid {
		fail(w, 400, "id inválido")
		return
	}
	m, err := s.Queries.Metrics(r.Context(), id)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, 200, m)
}

// ---- columnas ----

func (s *Server) addColumn(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	var req struct {
		Title string `json:"title"`
	}
	if !valid || !decode(r, &req) {
		fail(w, 400, "petición inválida")
		return
	}
	c, err := s.Boards.AddColumn(r.Context(), id, req.Title)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, 201, appkb.ColumnViewOf(c))
}

func (s *Server) updateColumn(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	m, valid2 := decodeMap(r)
	if !valid || !valid2 {
		fail(w, 400, "petición inválida")
		return
	}
	patch, err := parseColumnPatch(m)
	if err != nil {
		failErr(w, err)
		return
	}
	if err := s.Boards.UpdateColumn(r.Context(), id, patch); err != nil {
		failErr(w, err)
		return
	}
	ok(w)
}

func (s *Server) deleteColumn(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	if !valid {
		fail(w, 400, "id inválido")
		return
	}
	if err := s.Boards.DeleteColumn(r.Context(), id); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) reorderColumns(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	var req struct {
		IDs []int `json:"ids"`
	}
	if !valid || !decode(r, &req) {
		fail(w, 400, "petición inválida")
		return
	}
	if err := s.Boards.ReorderColumns(r.Context(), id, req.IDs); err != nil {
		failErr(w, err)
		return
	}
	ok(w)
}

// ---- etiquetas ----

type labelReq struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (s *Server) createLabel(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	var req labelReq
	if !valid || !decode(r, &req) {
		fail(w, 400, "nombre y color (#rrggbb) requeridos")
		return
	}
	l, err := s.Labels.Create(r.Context(), id, req.Name, req.Color)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, 201, appkb.LabelViewOf(l))
}

func (s *Server) updateLabel(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	var req labelReq
	if !valid || !decode(r, &req) {
		fail(w, 400, "nombre y color (#rrggbb) requeridos")
		return
	}
	if err := s.Labels.Update(r.Context(), id, req.Name, req.Color); err != nil {
		failErr(w, err)
		return
	}
	ok(w)
}

func (s *Server) deleteLabel(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	if !valid {
		fail(w, 400, "id inválido")
		return
	}
	if err := s.Labels.Delete(r.Context(), id); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- parseo de parches parciales (JSON → comandos de aplicación) ----

func parseColumnPatch(m map[string]any) (appkb.ColumnPatch, error) {
	var p appkb.ColumnPatch
	if v, present := m["title"]; present {
		t, isStr := v.(string)
		if !isStr {
			return p, shared.ErrInvalid
		}
		p.Title = &t
	}
	if v, present := m["wip_limit"]; present {
		p.WIPLimit.Present = true
		if v != nil {
			n, err := toInt(v)
			if err != nil {
				return p, err
			}
			p.WIPLimit.Value = &n
		}
	}
	if v, present := m["kind"]; present {
		k, isStr := v.(string)
		if !isStr {
			return p, shared.ErrInvalid
		}
		p.Kind = &k
	}
	return p, nil
}

func toInt(v any) (int, error) {
	f, isNum := v.(float64)
	if !isNum || f != float64(int(f)) {
		return 0, shared.ErrInvalid
	}
	return int(f), nil
}
