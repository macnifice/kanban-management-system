package httpapi

import (
	"net/http"

	appkb "kanban/internal/application/kanban"
	"kanban/internal/domain/shared"
)

func (s *Server) createCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ColumnID    int    `json:"column_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
		AssigneeID  int    `json:"assignee_id"`
		LabelIDs    []int  `json:"label_ids"`
	}
	if !decode(r, &req) || req.ColumnID == 0 {
		fail(w, 400, "column_id y título requeridos")
		return
	}
	c, err := s.Cards.Create(r.Context(), appkb.CreateCardCommand{
		ColumnID: req.ColumnID, Title: req.Title, Description: req.Description,
		Priority: req.Priority, AssigneeID: req.AssigneeID, LabelIDs: req.LabelIDs,
	}, userID(r))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, 201, appkb.CardViewOf(c))
}

func (s *Server) updateCard(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	m, valid2 := decodeMap(r)
	if !valid || !valid2 {
		fail(w, 400, "petición inválida")
		return
	}
	patch, err := parseCardPatch(m)
	if err != nil {
		failErr(w, err)
		return
	}
	c, err := s.Cards.Update(r.Context(), id, patch)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, 200, appkb.CardViewOf(c))
}

func (s *Server) deleteCard(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	if !valid {
		fail(w, 400, "id inválido")
		return
	}
	if err := s.Cards.Delete(r.Context(), id); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) moveCard(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	var req struct {
		ColumnID int `json:"column_id"`
		Position int `json:"position"`
	}
	if !valid || !decode(r, &req) || req.ColumnID == 0 {
		fail(w, 400, "petición inválida")
		return
	}
	if err := s.Cards.Move(r.Context(), id, req.ColumnID, req.Position, userID(r)); err != nil {
		failErr(w, err)
		return
	}
	ok(w)
}

func (s *Server) history(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r)
	if !valid {
		fail(w, 400, "id inválido")
		return
	}
	ev, err := s.Queries.CardHistory(r.Context(), id)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, 200, ev)
}

func parseCardPatch(m map[string]any) (appkb.CardPatch, error) {
	var p appkb.CardPatch
	str := func(key string) (*string, error) {
		v, present := m[key]
		if !present {
			return nil, nil
		}
		s, isStr := v.(string)
		if !isStr {
			return nil, shared.ErrInvalid
		}
		return &s, nil
	}
	var err error
	if p.Title, err = str("title"); err != nil {
		return p, err
	}
	if p.Description, err = str("description"); err != nil {
		return p, err
	}
	if p.Priority, err = str("priority"); err != nil {
		return p, err
	}
	if v, present := m["assignee_id"]; present {
		p.Assignee.Present = true
		if v != nil {
			n, err := toInt(v)
			if err != nil || n < 1 {
				return p, shared.ErrInvalid
			}
			p.Assignee.Value = &n
		}
	}
	if v, present := m["label_ids"]; present {
		arr, isArr := v.([]any)
		if !isArr {
			return p, shared.ErrInvalid
		}
		ids := []int{}
		for _, x := range arr {
			n, err := toInt(x)
			if err != nil {
				return p, err
			}
			ids = append(ids, n)
		}
		p.LabelIDs = &ids
	}
	return p, nil
}
