package handler

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peter/tacticarium/backend/internal/models"
)

type MissionHandler struct {
	db *pgxpool.Pool
}

func NewMissionHandler(db *pgxpool.Pool) *MissionHandler {
	return &MissionHandler{db: db}
}

// ListForceDispositions returns the five 11e force dispositions.
func (h *MissionHandler) ListForceDispositions(ctx context.Context, input *struct{}) (*ForceDispositionListOutput, error) {
	rows, err := h.db.Query(ctx, `SELECT id, name, text FROM force_dispositions ORDER BY name`)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	out := make([]models.ForceDisposition, 0)
	for rows.Next() {
		var d models.ForceDisposition
		if err := rows.Scan(&d.ID, &d.Name, &d.Text); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		out = append(out, d)
	}
	return &ForceDispositionListOutput{Body: out}, nil
}

// ListMissions returns the 11e primary mission objective records.
func (h *MissionHandler) ListMissions(ctx context.Context, input *struct{}) (*MissionListOutput, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, name, vp_per_game_cap, vp_per_round_cap, deployment_pattern_ids
		 FROM primary_missions ORDER BY name`)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	out := make([]models.Mission, 0)
	for rows.Next() {
		var m models.Mission
		if err := rows.Scan(&m.ID, &m.Name, &m.VPPerGameCap, &m.VPPerRoundCap, &m.DeploymentPatternIDs); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		out = append(out, m)
	}
	return &MissionListOutput{Body: out}, nil
}

// ListMissionMatchups returns the 5x5 disposition selector matrix.
func (h *MissionHandler) ListMissionMatchups(ctx context.Context, input *struct{}) (*MissionMatchupListOutput, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, disposition, opponent_disposition, mission_id FROM mission_matchups ORDER BY id`)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	out := make([]models.MissionMatchup, 0)
	for rows.Next() {
		var m models.MissionMatchup
		if err := rows.Scan(&m.ID, &m.Disposition, &m.OpponentDisposition, &m.MissionID); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		out = append(out, m)
	}
	return &MissionMatchupListOutput{Body: out}, nil
}

// ListSecondaryCards returns the secondary mission deck.
func (h *MissionHandler) ListSecondaryCards(ctx context.Context, input *struct{}) (*CardListOutput, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, name, card_type, COALESCE(subtype, ''), text
		 FROM cards WHERE card_type = 'secondary' ORDER BY name`)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	out := make([]models.MissionCard, 0)
	for rows.Next() {
		var c models.MissionCard
		if err := rows.Scan(&c.ID, &c.Name, &c.CardType, &c.Subtype, &c.Text); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		out = append(out, c)
	}
	return &CardListOutput{Body: out}, nil
}

// ListDeploymentPatterns returns the board deployment patterns.
func (h *MissionHandler) ListDeploymentPatterns(ctx context.Context, input *struct{}) (*DeploymentPatternListOutput, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, name, COALESCE(source, ''), COALESCE(description, ''),
		        objectives, territories, zones, recommended_terrain_layout_ids
		 FROM deployment_patterns ORDER BY name`)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	out := make([]models.DeploymentPattern, 0)
	for rows.Next() {
		var p models.DeploymentPattern
		if err := rows.Scan(&p.ID, &p.Name, &p.Source, &p.Description,
			&p.Objectives, &p.Territories, &p.Zones, &p.RecommendedTerrainLayoutIDs); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		out = append(out, p)
	}
	return &DeploymentPatternListOutput{Body: out}, nil
}
