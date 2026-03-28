package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peter/tacticarium/backend/internal/models"
)

type FactionHandler struct {
	db *pgxpool.Pool
}

func NewFactionHandler(db *pgxpool.Pool) *FactionHandler {
	return &FactionHandler{db: db}
}

func (h *FactionHandler) ListFactions(w http.ResponseWriter, r *http.Request) {
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

func (h *FactionHandler) ListDetachments(w http.ResponseWriter, r *http.Request) {
	factionID := chi.URLParam(r, "factionId")

	rows, err := h.db.Query(r.Context(),
		`SELECT id, faction_id, name FROM detachments WHERE faction_id = $1 ORDER BY name`, factionID)
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

func (h *FactionHandler) ListStratagems(w http.ResponseWriter, r *http.Request) {
	factionID := chi.URLParam(r, "factionId")

	rows, err := h.db.Query(r.Context(),
		`SELECT id, COALESCE(faction_id, ''), COALESCE(detachment_id, ''), name, type, cp_cost, COALESCE(legend, ''), turn, phase, description
		 FROM stratagems WHERE faction_id = $1 OR faction_id IS NULL ORDER BY name`, factionID)
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

func (h *FactionHandler) ListDetachmentStratagems(w http.ResponseWriter, r *http.Request) {
	detachmentID := chi.URLParam(r, "detachmentId")

	rows, err := h.db.Query(r.Context(),
		`SELECT id, COALESCE(faction_id, ''), COALESCE(detachment_id, ''), name, type, cp_cost, COALESCE(legend, ''), turn, phase, description
		 FROM stratagems WHERE detachment_id = $1 ORDER BY name`, detachmentID)
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
