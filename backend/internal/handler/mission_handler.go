package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peter/tacticarium/backend/internal/models"
)

type MissionHandler struct {
	db *pgxpool.Pool
}

func NewMissionHandler(db *pgxpool.Pool) *MissionHandler {
	return &MissionHandler{db: db}
}

func (h *MissionHandler) ListMissionPacks(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `SELECT id, name, COALESCE(description, '') FROM mission_packs ORDER BY name`)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var packs []models.MissionPack
	for rows.Next() {
		var p models.MissionPack
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		packs = append(packs, p)
	}

	writeJSON(w, http.StatusOK, packs)
}

func (h *MissionHandler) ListMissions(w http.ResponseWriter, r *http.Request) {
	packID := chi.URLParam(r, "packId")

	rows, err := h.db.Query(r.Context(),
		`SELECT id, mission_pack_id, name, lore, description
		 FROM missions WHERE mission_pack_id = $1 ORDER BY name`, packID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	missions := make([]models.Mission, 0)
	for rows.Next() {
		var m models.Mission
		if err := rows.Scan(&m.ID, &m.MissionPackID, &m.Name, &m.Lore, &m.Description); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		missions = append(missions, m)
	}

	writeJSON(w, http.StatusOK, missions)
}

func (h *MissionHandler) ListSecondaries(w http.ResponseWriter, r *http.Request) {
	packID := chi.URLParam(r, "packId")

	rows, err := h.db.Query(r.Context(),
		`SELECT id, mission_pack_id, name, lore, description, max_vp, is_fixed
		 FROM secondaries WHERE mission_pack_id = $1 ORDER BY name`, packID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	secondaries := make([]models.Secondary, 0)
	for rows.Next() {
		var s models.Secondary
		if err := rows.Scan(&s.ID, &s.MissionPackID, &s.Name, &s.Lore, &s.Description, &s.MaxVP, &s.IsFixed); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		secondaries = append(secondaries, s)
	}

	writeJSON(w, http.StatusOK, secondaries)
}

func (h *MissionHandler) ListMissionRules(w http.ResponseWriter, r *http.Request) {
	packID := chi.URLParam(r, "packId")

	rows, err := h.db.Query(r.Context(),
		`SELECT id, mission_pack_id, name, lore, description
		 FROM mission_rules WHERE mission_pack_id = $1 ORDER BY name`, packID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	rules := make([]models.MissionRule, 0)
	for rows.Next() {
		var mr models.MissionRule
		if err := rows.Scan(&mr.ID, &mr.MissionPackID, &mr.Name, &mr.Lore, &mr.Description); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		rules = append(rules, mr)
	}

	writeJSON(w, http.StatusOK, rules)
}

func (h *MissionHandler) ListChallengerCards(w http.ResponseWriter, r *http.Request) {
	packID := chi.URLParam(r, "packId")

	rows, err := h.db.Query(r.Context(),
		`SELECT id, mission_pack_id, name, lore, description
		 FROM challenger_cards WHERE mission_pack_id = $1 ORDER BY name`, packID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	cards := make([]models.ChallengerCard, 0)
	for rows.Next() {
		var c models.ChallengerCard
		if err := rows.Scan(&c.ID, &c.MissionPackID, &c.Name, &c.Lore, &c.Description); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		cards = append(cards, c)
	}

	writeJSON(w, http.StatusOK, cards)
}

func (h *MissionHandler) ListGambits(w http.ResponseWriter, r *http.Request) {
	packID := chi.URLParam(r, "packId")

	rows, err := h.db.Query(r.Context(),
		`SELECT id, mission_pack_id, name, description, vp_value
		 FROM gambits WHERE mission_pack_id = $1 ORDER BY name`, packID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var gambits []models.Gambit
	for rows.Next() {
		var g models.Gambit
		if err := rows.Scan(&g.ID, &g.MissionPackID, &g.Name, &g.Description, &g.VPValue); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		gambits = append(gambits, g)
	}

	writeJSON(w, http.StatusOK, gambits)
}
