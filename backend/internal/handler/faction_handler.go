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
	rows, err := h.db.Query(ctx,
		`SELECT id, name, COALESCE(parent_faction_id, ''), COALESCE(faction_rule_id, '')
		 FROM factions ORDER BY name`)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	factions := make([]models.Faction, 0)
	for rows.Next() {
		var f models.Faction
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentFactionID, &f.FactionRuleID); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		factions = append(factions, f)
	}
	return &FactionListOutput{Body: factions}, nil
}

func (h *FactionHandler) ListDetachments(ctx context.Context, input *FactionIDParam) (*DetachmentListOutput, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, faction_id, name, COALESCE(detachment_points, 0), force_dispositions
		 FROM detachments WHERE faction_id = $1 ORDER BY name`, input.FactionID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	detachments := make([]models.Detachment, 0)
	for rows.Next() {
		var d models.Detachment
		if err := rows.Scan(&d.ID, &d.FactionID, &d.Name, &d.DetachmentPoints, &d.ForceDispositions); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		detachments = append(detachments, d)
	}
	return &DetachmentListOutput{Body: detachments}, nil
}

const stratagemSelect = `SELECT id, COALESCE(faction_id, ''), COALESCE(detachment_id, ''), name,
	        COALESCE(type, ''), COALESCE(category, ''), cp_cost, phases,
	        COALESCE(player_turn, ''), COALESCE(timing, '')
	 FROM stratagems_11e`

func scanStratagems(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
},
) ([]models.Stratagem, error) {
	stratagems := make([]models.Stratagem, 0)
	for rows.Next() {
		var s models.Stratagem
		if err := rows.Scan(&s.ID, &s.FactionID, &s.DetachmentID, &s.Name,
			&s.Type, &s.Category, &s.CPCost, &s.Phases, &s.PlayerTurn, &s.Timing); err != nil {
			return nil, err
		}
		stratagems = append(stratagems, s)
	}
	return stratagems, rows.Err()
}

func (h *FactionHandler) ListStratagems(ctx context.Context, input *FactionIDParam) (*StratagemListOutput, error) {
	rows, err := h.db.Query(ctx,
		stratagemSelect+` WHERE (faction_id = $1 OR faction_id IS NULL) ORDER BY name`, input.FactionID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	stratagems, err := scanStratagems(rows)
	if err != nil {
		return nil, huma.Error500InternalServerError("scan error")
	}
	return &StratagemListOutput{Body: stratagems}, nil
}

func (h *FactionHandler) ListDetachmentStratagems(ctx context.Context, input *DetachmentIDParam) (*StratagemListOutput, error) {
	rows, err := h.db.Query(ctx,
		stratagemSelect+` WHERE detachment_id = $1 ORDER BY name`, input.DetachmentID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	stratagems, err := scanStratagems(rows)
	if err != nil {
		return nil, huma.Error500InternalServerError("scan error")
	}
	return &StratagemListOutput{Body: stratagems}, nil
}
