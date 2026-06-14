package handler

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
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

func (h *AdminHandler) ListFactions(ctx context.Context, input *struct{}) (*FactionListOutput, error) {
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

func (h *AdminHandler) GetFaction(ctx context.Context, input *IDParam) (*FactionOutput, error) {
	var f models.Faction
	err := h.db.QueryRow(ctx,
		`SELECT id, name, COALESCE(parent_faction_id, ''), COALESCE(faction_rule_id, '')
		 FROM factions WHERE id = $1`, input.ID).
		Scan(&f.ID, &f.Name, &f.ParentFactionID, &f.FactionRuleID)
	if err != nil {
		return nil, huma.Error404NotFound("not found")
	}
	return &FactionOutput{Body: f}, nil
}

func (h *AdminHandler) CreateFaction(ctx context.Context, input *FactionInput) (*FactionOutput, error) {
	f := input.Body
	if f.ID == "" || f.Name == "" {
		return nil, huma.Error400BadRequest("id and name are required")
	}
	_, err := h.db.Exec(ctx,
		`INSERT INTO factions (id, name, parent_faction_id, faction_rule_id)
		 VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''))`,
		f.ID, f.Name, f.ParentFactionID, f.FactionRuleID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	return &FactionOutput{Body: f}, nil
}

func (h *AdminHandler) UpdateFaction(ctx context.Context, input *IDFactionInput) (*FactionOutput, error) {
	f := input.Body
	tag, err := h.db.Exec(ctx,
		`UPDATE factions SET name = $1, parent_faction_id = NULLIF($2, ''), faction_rule_id = NULLIF($3, '')
		 WHERE id = $4`,
		f.Name, f.ParentFactionID, f.FactionRuleID, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	f.ID = input.ID
	return &FactionOutput{Body: f}, nil
}

func (h *AdminHandler) DeleteFaction(ctx context.Context, input *IDParam) (*struct{}, error) {
	tag, err := h.db.Exec(ctx, `DELETE FROM factions WHERE id = $1`, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	return nil, nil
}

// --- Detachments ---

func (h *AdminHandler) ListDetachments(ctx context.Context, input *AdminDetachmentListInput) (*DetachmentListOutput, error) {
	query := `SELECT id, faction_id, name, COALESCE(detachment_points, 0), force_dispositions FROM detachments`
	var args []any
	if input.FactionID != "" {
		query += ` WHERE faction_id = $1 ORDER BY name`
		args = []any{input.FactionID}
	} else {
		query += ` ORDER BY faction_id, name`
	}
	rows, err := h.db.Query(ctx, query, args...)
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

func (h *AdminHandler) GetDetachment(ctx context.Context, input *IDParam) (*DetachmentOutput, error) {
	var d models.Detachment
	err := h.db.QueryRow(ctx,
		`SELECT id, faction_id, name, COALESCE(detachment_points, 0), force_dispositions
		 FROM detachments WHERE id = $1`, input.ID).
		Scan(&d.ID, &d.FactionID, &d.Name, &d.DetachmentPoints, &d.ForceDispositions)
	if err != nil {
		return nil, huma.Error404NotFound("not found")
	}
	return &DetachmentOutput{Body: d}, nil
}

func (h *AdminHandler) CreateDetachment(ctx context.Context, input *DetachmentInput) (*DetachmentOutput, error) {
	d := input.Body
	if d.ID == "" || d.FactionID == "" || d.Name == "" {
		return nil, huma.Error400BadRequest("id, factionId and name are required")
	}
	if d.ForceDispositions == nil {
		d.ForceDispositions = []string{}
	}
	_, err := h.db.Exec(ctx,
		`INSERT INTO detachments (id, faction_id, name, detachment_points, force_dispositions)
		 VALUES ($1, $2, $3, NULLIF($4, 0), $5)`,
		d.ID, d.FactionID, d.Name, d.DetachmentPoints, d.ForceDispositions)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	return &DetachmentOutput{Body: d}, nil
}

func (h *AdminHandler) UpdateDetachment(ctx context.Context, input *IDDetachmentInput) (*DetachmentOutput, error) {
	d := input.Body
	if d.ForceDispositions == nil {
		d.ForceDispositions = []string{}
	}
	tag, err := h.db.Exec(ctx,
		`UPDATE detachments SET faction_id = $1, name = $2, detachment_points = NULLIF($3, 0),
		 force_dispositions = $4 WHERE id = $5`,
		d.FactionID, d.Name, d.DetachmentPoints, d.ForceDispositions, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	d.ID = input.ID
	return &DetachmentOutput{Body: d}, nil
}

func (h *AdminHandler) DeleteDetachment(ctx context.Context, input *IDParam) (*struct{}, error) {
	tag, err := h.db.Exec(ctx, `DELETE FROM detachments WHERE id = $1`, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	return nil, nil
}

// --- Stratagems (read-only: bulk-seeded from 40kdc-data, detachment-scoped) ---

func (h *AdminHandler) ListStratagems(ctx context.Context, input *AdminStratagemListInput) (*StratagemListOutput, error) {
	query := stratagemSelect
	var args []any
	switch {
	case input.DetachmentID != "":
		query += ` WHERE detachment_id = $1 ORDER BY name`
		args = []any{input.DetachmentID}
	case input.FactionID != "":
		query += ` WHERE faction_id = $1 ORDER BY name`
		args = []any{input.FactionID}
	default:
		query += ` ORDER BY faction_id, name`
	}
	rows, err := h.db.Query(ctx, query, args...)
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
