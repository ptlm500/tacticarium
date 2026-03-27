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
		`SELECT id, mission_pack_id, name, COALESCE(description, ''), COALESCE(deployment_map, ''), COALESCE(rules_text, '')
		 FROM missions WHERE mission_pack_id = $1 ORDER BY name`, packID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var missions []models.Mission
	for rows.Next() {
		var m models.Mission
		if err := rows.Scan(&m.ID, &m.MissionPackID, &m.Name, &m.Description, &m.DeploymentMap, &m.RulesText); err != nil {
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
		`SELECT id, mission_pack_id, name, category, description, max_vp
		 FROM secondaries WHERE mission_pack_id = $1 ORDER BY category, name`, packID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var secondaries []models.Secondary
	for rows.Next() {
		var s models.Secondary
		if err := rows.Scan(&s.ID, &s.MissionPackID, &s.Name, &s.Category, &s.Description, &s.MaxVP); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		secondaries = append(secondaries, s)
	}

	writeJSON(w, http.StatusOK, secondaries)
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
