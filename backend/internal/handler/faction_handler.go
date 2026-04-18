package handler

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peter/tacticarium/backend/internal/models"
)

type FactionHandler struct {
	db *pgxpool.Pool
}

func NewFactionHandler(db *pgxpool.Pool) *FactionHandler {
	return &FactionHandler{db: db}
}

func (h *FactionHandler) ListFactions(ctx context.Context, input *struct{}) (*FactionListOutput, error) {
	rows, err := h.db.Query(ctx, `SELECT id, name, COALESCE(wahapedia_link, '') FROM factions ORDER BY name`)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	factions := make([]models.Faction, 0)
	for rows.Next() {
		var f models.Faction
		if err := rows.Scan(&f.ID, &f.Name, &f.WahapediaLink); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		factions = append(factions, f)
	}

	return &FactionListOutput{Body: factions}, nil
}

func (h *FactionHandler) ListDetachments(ctx context.Context, input *FactionIDParam) (*DetachmentListOutput, error) {
	// Only the "core" game mode is currently supported — detachments exclusive
	// to alternate modes (e.g. Boarding Actions) are hidden from players.
	rows, err := h.db.Query(ctx,
		`SELECT id, faction_id, name, game_mode FROM detachments WHERE faction_id = $1 AND game_mode = 'core' ORDER BY name`, input.FactionID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	detachments := make([]models.Detachment, 0)
	for rows.Next() {
		var d models.Detachment
		if err := rows.Scan(&d.ID, &d.FactionID, &d.Name, &d.GameMode); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		detachments = append(detachments, d)
	}

	return &DetachmentListOutput{Body: detachments}, nil
}

func (h *FactionHandler) ListStratagems(ctx context.Context, input *FactionIDParam) (*StratagemListOutput, error) {
	// Only the "core" game mode is currently supported — stratagems belonging
	// to alternate modes (e.g. Boarding Actions) are hidden from players.
	rows, err := h.db.Query(ctx,
		`SELECT id, COALESCE(faction_id, ''), COALESCE(detachment_id, ''), name, type, cp_cost, COALESCE(legend, ''), turn, phase, description, game_mode
		 FROM stratagems WHERE (faction_id = $1 OR faction_id IS NULL) AND game_mode = 'core' ORDER BY name`, input.FactionID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	stratagems := make([]models.Stratagem, 0)
	for rows.Next() {
		var s models.Stratagem
		if err := rows.Scan(&s.ID, &s.FactionID, &s.DetachmentID, &s.Name, &s.Type, &s.CPCost, &s.Legend, &s.Turn, &s.Phase, &s.Description, &s.GameMode); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		stratagems = append(stratagems, s)
	}

	return &StratagemListOutput{Body: stratagems}, nil
}

func (h *FactionHandler) ListDetachmentStratagems(ctx context.Context, input *DetachmentIDParam) (*StratagemListOutput, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, COALESCE(faction_id, ''), COALESCE(detachment_id, ''), name, type, cp_cost, COALESCE(legend, ''), turn, phase, description, game_mode
		 FROM stratagems WHERE detachment_id = $1 AND game_mode = 'core' ORDER BY name`, input.DetachmentID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	stratagems := make([]models.Stratagem, 0)
	for rows.Next() {
		var s models.Stratagem
		if err := rows.Scan(&s.ID, &s.FactionID, &s.DetachmentID, &s.Name, &s.Type, &s.CPCost, &s.Legend, &s.Turn, &s.Phase, &s.Description, &s.GameMode); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		stratagems = append(stratagems, s)
	}

	return &StratagemListOutput{Body: stratagems}, nil
}
