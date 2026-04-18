package handler

import (
	"context"
	"encoding/json"

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

func (h *MissionHandler) ListMissionPacks(ctx context.Context, input *struct{}) (*MissionPackListOutput, error) {
	rows, err := h.db.Query(ctx, `SELECT id, name, COALESCE(description, '') FROM mission_packs ORDER BY name`)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	packs := make([]models.MissionPack, 0)
	for rows.Next() {
		var p models.MissionPack
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		packs = append(packs, p)
	}

	return &MissionPackListOutput{Body: packs}, nil
}

func (h *MissionHandler) ListMissions(ctx context.Context, input *PackIDParam) (*MissionListOutput, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, mission_pack_id, name, lore, description, scoring_rules, scoring_timing
		 FROM missions WHERE mission_pack_id = $1 ORDER BY name`, input.PackID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	missions := make([]models.Mission, 0)
	for rows.Next() {
		var m models.Mission
		var scoringJSON []byte
		if err := rows.Scan(&m.ID, &m.MissionPackID, &m.Name, &m.Lore, &m.Description, &scoringJSON, &m.ScoringTiming); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		json.Unmarshal(scoringJSON, &m.ScoringRules)
		if m.ScoringRules == nil {
			m.ScoringRules = []models.ScoringAction{}
		}
		missions = append(missions, m)
	}

	return &MissionListOutput{Body: missions}, nil
}

func (h *MissionHandler) ListSecondaries(ctx context.Context, input *PackIDParam) (*SecondaryListOutput, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, mission_pack_id, name, lore, description, max_vp, is_fixed, scoring_options, draw_restriction
		 FROM secondaries WHERE mission_pack_id = $1 ORDER BY name`, input.PackID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	secondaries := make([]models.Secondary, 0)
	for rows.Next() {
		var s models.Secondary
		var optionsJSON []byte
		var drawJSON []byte
		if err := rows.Scan(&s.ID, &s.MissionPackID, &s.Name, &s.Lore, &s.Description, &s.MaxVP, &s.IsFixed, &optionsJSON, &drawJSON); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		json.Unmarshal(optionsJSON, &s.ScoringOptions)
		if s.ScoringOptions == nil {
			s.ScoringOptions = []models.ScoringOption{}
		}
		if len(drawJSON) > 0 {
			var dr models.DrawRestriction
			if err := json.Unmarshal(drawJSON, &dr); err == nil {
				s.DrawRestriction = &dr
			}
		}
		secondaries = append(secondaries, s)
	}

	return &SecondaryListOutput{Body: secondaries}, nil
}

func (h *MissionHandler) ListMissionRules(ctx context.Context, input *PackIDParam) (*MissionRuleListOutput, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, mission_pack_id, name, lore, description
		 FROM mission_rules WHERE mission_pack_id = $1 ORDER BY name`, input.PackID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	rules := make([]models.MissionRule, 0)
	for rows.Next() {
		var mr models.MissionRule
		if err := rows.Scan(&mr.ID, &mr.MissionPackID, &mr.Name, &mr.Lore, &mr.Description); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		rules = append(rules, mr)
	}

	return &MissionRuleListOutput{Body: rules}, nil
}

func (h *MissionHandler) ListChallengerCards(ctx context.Context, input *PackIDParam) (*ChallengerCardListOutput, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, mission_pack_id, name, lore, description
		 FROM challenger_cards WHERE mission_pack_id = $1 ORDER BY name`, input.PackID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	cards := make([]models.ChallengerCard, 0)
	for rows.Next() {
		var c models.ChallengerCard
		if err := rows.Scan(&c.ID, &c.MissionPackID, &c.Name, &c.Lore, &c.Description); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		cards = append(cards, c)
	}

	return &ChallengerCardListOutput{Body: cards}, nil
}

func (h *MissionHandler) ListGambits(ctx context.Context, input *PackIDParam) (*GambitListOutput, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, mission_pack_id, name, description, vp_value
		 FROM gambits WHERE mission_pack_id = $1 ORDER BY name`, input.PackID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	gambits := make([]models.Gambit, 0)
	for rows.Next() {
		var g models.Gambit
		if err := rows.Scan(&g.ID, &g.MissionPackID, &g.Name, &g.Description, &g.VPValue); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		gambits = append(gambits, g)
	}

	return &GambitListOutput{Body: gambits}, nil
}
