package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peter/tacticarium/backend/internal/models"
)

type AdminHandler struct {
	db *pgxpool.Pool
}

func NewAdminHandler(db *pgxpool.Pool) *AdminHandler {
	return &AdminHandler{db: db}
}

// --- Factions ---

func (h *AdminHandler) ListFactions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `SELECT id, name, COALESCE(wahapedia_link, '') FROM factions ORDER BY name`)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var factions []models.Faction
	for rows.Next() {
		var f models.Faction
		if err := rows.Scan(&f.ID, &f.Name, &f.WahapediaLink); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		factions = append(factions, f)
	}
	writeJSON(w, http.StatusOK, factions)
}

func (h *AdminHandler) GetFaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var f models.Faction
	err := h.db.QueryRow(r.Context(), `SELECT id, name, COALESCE(wahapedia_link, '') FROM factions WHERE id = $1`, id).
		Scan(&f.ID, &f.Name, &f.WahapediaLink)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *AdminHandler) CreateFaction(w http.ResponseWriter, r *http.Request) {
	var f models.Faction
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if f.ID == "" || f.Name == "" {
		http.Error(w, "id and name are required", http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec(r.Context(),
		`INSERT INTO factions (id, name, wahapedia_link) VALUES ($1, $2, NULLIF($3, ''))`,
		f.ID, f.Name, f.WahapediaLink)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

func (h *AdminHandler) UpdateFaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var f models.Faction
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	tag, err := h.db.Exec(r.Context(),
		`UPDATE factions SET name = $1, wahapedia_link = NULLIF($2, '') WHERE id = $3`,
		f.Name, f.WahapediaLink, id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	f.ID = id
	writeJSON(w, http.StatusOK, f)
}

func (h *AdminHandler) DeleteFaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag, err := h.db.Exec(r.Context(), `DELETE FROM factions WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Detachments ---

func (h *AdminHandler) ListDetachments(w http.ResponseWriter, r *http.Request) {
	factionID := r.URL.Query().Get("faction_id")
	var query string
	var args []any
	if factionID != "" {
		query = `SELECT id, faction_id, name FROM detachments WHERE faction_id = $1 ORDER BY name`
		args = []any{factionID}
	} else {
		query = `SELECT id, faction_id, name FROM detachments ORDER BY faction_id, name`
	}
	rows, err := h.db.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var detachments []models.Detachment
	for rows.Next() {
		var d models.Detachment
		if err := rows.Scan(&d.ID, &d.FactionID, &d.Name); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		detachments = append(detachments, d)
	}
	writeJSON(w, http.StatusOK, detachments)
}

func (h *AdminHandler) GetDetachment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var d models.Detachment
	err := h.db.QueryRow(r.Context(), `SELECT id, faction_id, name FROM detachments WHERE id = $1`, id).
		Scan(&d.ID, &d.FactionID, &d.Name)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *AdminHandler) CreateDetachment(w http.ResponseWriter, r *http.Request) {
	var d models.Detachment
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if d.ID == "" || d.FactionID == "" || d.Name == "" {
		http.Error(w, "id, factionId, and name are required", http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec(r.Context(),
		`INSERT INTO detachments (id, faction_id, name) VALUES ($1, $2, $3)`,
		d.ID, d.FactionID, d.Name)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, d)
}

func (h *AdminHandler) UpdateDetachment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var d models.Detachment
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	tag, err := h.db.Exec(r.Context(),
		`UPDATE detachments SET faction_id = $1, name = $2 WHERE id = $3`,
		d.FactionID, d.Name, id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	d.ID = id
	writeJSON(w, http.StatusOK, d)
}

func (h *AdminHandler) DeleteDetachment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag, err := h.db.Exec(r.Context(), `DELETE FROM detachments WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Stratagems ---

func (h *AdminHandler) ListStratagems(w http.ResponseWriter, r *http.Request) {
	factionID := r.URL.Query().Get("faction_id")
	detachmentID := r.URL.Query().Get("detachment_id")

	query := `SELECT id, COALESCE(faction_id, ''), COALESCE(detachment_id, ''), name, type, cp_cost, COALESCE(legend, ''), turn, phase, description FROM stratagems`
	var args []any
	var conditions []string

	if factionID != "" {
		args = append(args, factionID)
		conditions = append(conditions, fmt.Sprintf("faction_id = $%d", len(args)))
	}
	if detachmentID != "" {
		args = append(args, detachmentID)
		conditions = append(conditions, fmt.Sprintf("detachment_id = $%d", len(args)))
	}

	if len(conditions) > 0 {
		query += " WHERE "
		for i, c := range conditions {
			if i > 0 {
				query += " AND "
			}
			query += c
		}
	}
	query += " ORDER BY name"

	rows, err := h.db.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stratagems []models.Stratagem
	for rows.Next() {
		var s models.Stratagem
		if err := rows.Scan(&s.ID, &s.FactionID, &s.DetachmentID, &s.Name, &s.Type, &s.CPCost, &s.Legend, &s.Turn, &s.Phase, &s.Description); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		stratagems = append(stratagems, s)
	}
	writeJSON(w, http.StatusOK, stratagems)
}

func (h *AdminHandler) GetStratagem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var s models.Stratagem
	err := h.db.QueryRow(r.Context(),
		`SELECT id, COALESCE(faction_id, ''), COALESCE(detachment_id, ''), name, type, cp_cost, COALESCE(legend, ''), turn, phase, description FROM stratagems WHERE id = $1`, id).
		Scan(&s.ID, &s.FactionID, &s.DetachmentID, &s.Name, &s.Type, &s.CPCost, &s.Legend, &s.Turn, &s.Phase, &s.Description)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *AdminHandler) CreateStratagem(w http.ResponseWriter, r *http.Request) {
	var s models.Stratagem
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if s.ID == "" || s.Name == "" {
		http.Error(w, "id and name are required", http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec(r.Context(),
		`INSERT INTO stratagems (id, faction_id, detachment_id, name, type, cp_cost, legend, turn, phase, description)
		 VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), $4, $5, $6, NULLIF($7, ''), $8, $9, $10)`,
		s.ID, s.FactionID, s.DetachmentID, s.Name, s.Type, s.CPCost, s.Legend, s.Turn, s.Phase, s.Description)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *AdminHandler) UpdateStratagem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var s models.Stratagem
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	tag, err := h.db.Exec(r.Context(),
		`UPDATE stratagems SET faction_id = NULLIF($1, ''), detachment_id = NULLIF($2, ''), name = $3, type = $4, cp_cost = $5, legend = NULLIF($6, ''), turn = $7, phase = $8, description = $9 WHERE id = $10`,
		s.FactionID, s.DetachmentID, s.Name, s.Type, s.CPCost, s.Legend, s.Turn, s.Phase, s.Description, id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	s.ID = id
	writeJSON(w, http.StatusOK, s)
}

func (h *AdminHandler) DeleteStratagem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag, err := h.db.Exec(r.Context(), `DELETE FROM stratagems WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Mission Packs ---

func (h *AdminHandler) ListMissionPacks(w http.ResponseWriter, r *http.Request) {
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

func (h *AdminHandler) CreateMissionPack(w http.ResponseWriter, r *http.Request) {
	var p models.MissionPack
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if p.ID == "" || p.Name == "" {
		http.Error(w, "id and name are required", http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec(r.Context(),
		`INSERT INTO mission_packs (id, name, description) VALUES ($1, $2, NULLIF($3, ''))`,
		p.ID, p.Name, p.Description)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *AdminHandler) UpdateMissionPack(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var p models.MissionPack
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	tag, err := h.db.Exec(r.Context(),
		`UPDATE mission_packs SET name = $1, description = NULLIF($2, '') WHERE id = $3`,
		p.Name, p.Description, id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	p.ID = id
	writeJSON(w, http.StatusOK, p)
}

func (h *AdminHandler) DeleteMissionPack(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag, err := h.db.Exec(r.Context(), `DELETE FROM mission_packs WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Missions ---

func (h *AdminHandler) ListMissions(w http.ResponseWriter, r *http.Request) {
	packID := r.URL.Query().Get("pack_id")
	var query string
	var args []any
	if packID != "" {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, ''), COALESCE(scoring_rules, '[]'), COALESCE(scoring_timing, 'end_of_command_phase') FROM missions WHERE mission_pack_id = $1 ORDER BY name`
		args = []any{packID}
	} else {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, ''), COALESCE(scoring_rules, '[]'), COALESCE(scoring_timing, 'end_of_command_phase') FROM missions ORDER BY mission_pack_id, name`
	}
	rows, err := h.db.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var missions []models.Mission
	for rows.Next() {
		var m models.Mission
		var scoringRulesJSON string
		if err := rows.Scan(&m.ID, &m.MissionPackID, &m.Name, &m.Lore, &m.Description, &scoringRulesJSON, &m.ScoringTiming); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		json.Unmarshal([]byte(scoringRulesJSON), &m.ScoringRules)
		if m.ScoringRules == nil {
			m.ScoringRules = []models.ScoringAction{}
		}
		missions = append(missions, m)
	}
	writeJSON(w, http.StatusOK, missions)
}

func (h *AdminHandler) GetMission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.Mission
	var scoringRulesJSON string
	err := h.db.QueryRow(r.Context(),
		`SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, ''), COALESCE(scoring_rules, '[]'), COALESCE(scoring_timing, 'end_of_command_phase') FROM missions WHERE id = $1`, id).
		Scan(&m.ID, &m.MissionPackID, &m.Name, &m.Lore, &m.Description, &scoringRulesJSON, &m.ScoringTiming)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.Unmarshal([]byte(scoringRulesJSON), &m.ScoringRules)
	if m.ScoringRules == nil {
		m.ScoringRules = []models.ScoringAction{}
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *AdminHandler) CreateMission(w http.ResponseWriter, r *http.Request) {
	var m models.Mission
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if m.ID == "" || m.MissionPackID == "" || m.Name == "" {
		http.Error(w, "id, missionPackId, and name are required", http.StatusBadRequest)
		return
	}
	if m.ScoringRules == nil {
		m.ScoringRules = []models.ScoringAction{}
	}
	if m.ScoringTiming == "" {
		m.ScoringTiming = "end_of_command_phase"
	}
	scoringJSON, _ := json.Marshal(m.ScoringRules)
	_, err := h.db.Exec(r.Context(),
		`INSERT INTO missions (id, mission_pack_id, name, lore, description, scoring_rules, scoring_timing) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		m.ID, m.MissionPackID, m.Name, m.Lore, m.Description, string(scoringJSON), m.ScoringTiming)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *AdminHandler) UpdateMission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.Mission
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if m.ScoringRules == nil {
		m.ScoringRules = []models.ScoringAction{}
	}
	if m.ScoringTiming == "" {
		m.ScoringTiming = "end_of_command_phase"
	}
	scoringJSON, _ := json.Marshal(m.ScoringRules)
	tag, err := h.db.Exec(r.Context(),
		`UPDATE missions SET mission_pack_id = $1, name = $2, lore = $3, description = $4, scoring_rules = $5, scoring_timing = $6 WHERE id = $7`,
		m.MissionPackID, m.Name, m.Lore, m.Description, string(scoringJSON), m.ScoringTiming, id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	m.ID = id
	writeJSON(w, http.StatusOK, m)
}

func (h *AdminHandler) DeleteMission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag, err := h.db.Exec(r.Context(), `DELETE FROM missions WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Secondaries ---

func (h *AdminHandler) ListSecondaries(w http.ResponseWriter, r *http.Request) {
	packID := r.URL.Query().Get("pack_id")
	var query string
	var args []any
	if packID != "" {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, ''), max_vp, is_fixed, COALESCE(scoring_options, '[]') FROM secondaries WHERE mission_pack_id = $1 ORDER BY name`
		args = []any{packID}
	} else {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, ''), max_vp, is_fixed, COALESCE(scoring_options, '[]') FROM secondaries ORDER BY mission_pack_id, name`
	}
	rows, err := h.db.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var secondaries []models.Secondary
	for rows.Next() {
		var s models.Secondary
		var scoringJSON string
		if err := rows.Scan(&s.ID, &s.MissionPackID, &s.Name, &s.Lore, &s.Description, &s.MaxVP, &s.IsFixed, &scoringJSON); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		json.Unmarshal([]byte(scoringJSON), &s.ScoringOptions)
		if s.ScoringOptions == nil {
			s.ScoringOptions = []models.ScoringOption{}
		}
		secondaries = append(secondaries, s)
	}
	writeJSON(w, http.StatusOK, secondaries)
}

func (h *AdminHandler) GetSecondary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var s models.Secondary
	var scoringJSON string
	err := h.db.QueryRow(r.Context(),
		`SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, ''), max_vp, is_fixed, COALESCE(scoring_options, '[]') FROM secondaries WHERE id = $1`, id).
		Scan(&s.ID, &s.MissionPackID, &s.Name, &s.Lore, &s.Description, &s.MaxVP, &s.IsFixed, &scoringJSON)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.Unmarshal([]byte(scoringJSON), &s.ScoringOptions)
	if s.ScoringOptions == nil {
		s.ScoringOptions = []models.ScoringOption{}
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *AdminHandler) CreateSecondary(w http.ResponseWriter, r *http.Request) {
	var s models.Secondary
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if s.ID == "" || s.MissionPackID == "" || s.Name == "" {
		http.Error(w, "id, missionPackId, and name are required", http.StatusBadRequest)
		return
	}
	if s.ScoringOptions == nil {
		s.ScoringOptions = []models.ScoringOption{}
	}
	scoringJSON, _ := json.Marshal(s.ScoringOptions)
	_, err := h.db.Exec(r.Context(),
		`INSERT INTO secondaries (id, mission_pack_id, name, lore, description, max_vp, is_fixed, scoring_options) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		s.ID, s.MissionPackID, s.Name, s.Lore, s.Description, s.MaxVP, s.IsFixed, string(scoringJSON))
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *AdminHandler) UpdateSecondary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var s models.Secondary
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if s.ScoringOptions == nil {
		s.ScoringOptions = []models.ScoringOption{}
	}
	scoringJSON, _ := json.Marshal(s.ScoringOptions)
	tag, err := h.db.Exec(r.Context(),
		`UPDATE secondaries SET mission_pack_id = $1, name = $2, lore = $3, description = $4, max_vp = $5, is_fixed = $6, scoring_options = $7 WHERE id = $8`,
		s.MissionPackID, s.Name, s.Lore, s.Description, s.MaxVP, s.IsFixed, string(scoringJSON), id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	s.ID = id
	writeJSON(w, http.StatusOK, s)
}

func (h *AdminHandler) DeleteSecondary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag, err := h.db.Exec(r.Context(), `DELETE FROM secondaries WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Gambits ---

func (h *AdminHandler) ListGambits(w http.ResponseWriter, r *http.Request) {
	packID := r.URL.Query().Get("pack_id")
	var query string
	var args []any
	if packID != "" {
		query = `SELECT id, mission_pack_id, name, COALESCE(description, ''), vp_value FROM gambits WHERE mission_pack_id = $1 ORDER BY name`
		args = []any{packID}
	} else {
		query = `SELECT id, mission_pack_id, name, COALESCE(description, ''), vp_value FROM gambits ORDER BY mission_pack_id, name`
	}
	rows, err := h.db.Query(r.Context(), query, args...)
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

func (h *AdminHandler) GetGambit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var g models.Gambit
	err := h.db.QueryRow(r.Context(),
		`SELECT id, mission_pack_id, name, COALESCE(description, ''), vp_value FROM gambits WHERE id = $1`, id).
		Scan(&g.ID, &g.MissionPackID, &g.Name, &g.Description, &g.VPValue)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, g)
}

func (h *AdminHandler) CreateGambit(w http.ResponseWriter, r *http.Request) {
	var g models.Gambit
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if g.ID == "" || g.MissionPackID == "" || g.Name == "" {
		http.Error(w, "id, missionPackId, and name are required", http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec(r.Context(),
		`INSERT INTO gambits (id, mission_pack_id, name, description, vp_value) VALUES ($1, $2, $3, $4, $5)`,
		g.ID, g.MissionPackID, g.Name, g.Description, g.VPValue)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, g)
}

func (h *AdminHandler) UpdateGambit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var g models.Gambit
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	tag, err := h.db.Exec(r.Context(),
		`UPDATE gambits SET mission_pack_id = $1, name = $2, description = $3, vp_value = $4 WHERE id = $5`,
		g.MissionPackID, g.Name, g.Description, g.VPValue, id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	g.ID = id
	writeJSON(w, http.StatusOK, g)
}

func (h *AdminHandler) DeleteGambit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag, err := h.db.Exec(r.Context(), `DELETE FROM gambits WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Challenger Cards ---

func (h *AdminHandler) ListChallengerCards(w http.ResponseWriter, r *http.Request) {
	packID := r.URL.Query().Get("pack_id")
	var query string
	var args []any
	if packID != "" {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, '') FROM challenger_cards WHERE mission_pack_id = $1 ORDER BY name`
		args = []any{packID}
	} else {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, '') FROM challenger_cards ORDER BY mission_pack_id, name`
	}
	rows, err := h.db.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var cards []models.ChallengerCard
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

func (h *AdminHandler) GetChallengerCard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c models.ChallengerCard
	err := h.db.QueryRow(r.Context(),
		`SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, '') FROM challenger_cards WHERE id = $1`, id).
		Scan(&c.ID, &c.MissionPackID, &c.Name, &c.Lore, &c.Description)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *AdminHandler) CreateChallengerCard(w http.ResponseWriter, r *http.Request) {
	var c models.ChallengerCard
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if c.ID == "" || c.MissionPackID == "" || c.Name == "" {
		http.Error(w, "id, missionPackId, and name are required", http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec(r.Context(),
		`INSERT INTO challenger_cards (id, mission_pack_id, name, lore, description) VALUES ($1, $2, $3, $4, $5)`,
		c.ID, c.MissionPackID, c.Name, c.Lore, c.Description)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *AdminHandler) UpdateChallengerCard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c models.ChallengerCard
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	tag, err := h.db.Exec(r.Context(),
		`UPDATE challenger_cards SET mission_pack_id = $1, name = $2, lore = $3, description = $4 WHERE id = $5`,
		c.MissionPackID, c.Name, c.Lore, c.Description, id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	c.ID = id
	writeJSON(w, http.StatusOK, c)
}

func (h *AdminHandler) DeleteChallengerCard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag, err := h.db.Exec(r.Context(), `DELETE FROM challenger_cards WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Mission Rules ---

func (h *AdminHandler) ListMissionRules(w http.ResponseWriter, r *http.Request) {
	packID := r.URL.Query().Get("pack_id")
	var query string
	var args []any
	if packID != "" {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, '') FROM mission_rules WHERE mission_pack_id = $1 ORDER BY name`
		args = []any{packID}
	} else {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, '') FROM mission_rules ORDER BY mission_pack_id, name`
	}
	rows, err := h.db.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rules []models.MissionRule
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

func (h *AdminHandler) GetMissionRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var mr models.MissionRule
	err := h.db.QueryRow(r.Context(),
		`SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, '') FROM mission_rules WHERE id = $1`, id).
		Scan(&mr.ID, &mr.MissionPackID, &mr.Name, &mr.Lore, &mr.Description)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, mr)
}

func (h *AdminHandler) CreateMissionRule(w http.ResponseWriter, r *http.Request) {
	var mr models.MissionRule
	if err := json.NewDecoder(r.Body).Decode(&mr); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if mr.ID == "" || mr.MissionPackID == "" || mr.Name == "" {
		http.Error(w, "id, missionPackId, and name are required", http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec(r.Context(),
		`INSERT INTO mission_rules (id, mission_pack_id, name, lore, description) VALUES ($1, $2, $3, $4, $5)`,
		mr.ID, mr.MissionPackID, mr.Name, mr.Lore, mr.Description)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, mr)
}

func (h *AdminHandler) UpdateMissionRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var mr models.MissionRule
	if err := json.NewDecoder(r.Body).Decode(&mr); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	tag, err := h.db.Exec(r.Context(),
		`UPDATE mission_rules SET mission_pack_id = $1, name = $2, lore = $3, description = $4 WHERE id = $5`,
		mr.MissionPackID, mr.Name, mr.Lore, mr.Description, id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	mr.ID = id
	writeJSON(w, http.StatusOK, mr)
}

func (h *AdminHandler) DeleteMissionRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag, err := h.db.Exec(r.Context(), `DELETE FROM mission_rules WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
