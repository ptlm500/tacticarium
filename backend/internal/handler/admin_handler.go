package handler

import (
	"context"
	"encoding/json"
	"fmt"

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

func (h *AdminHandler) GetFaction(ctx context.Context, input *IDParam) (*FactionOutput, error) {
	var f models.Faction
	err := h.db.QueryRow(ctx, `SELECT id, name, COALESCE(wahapedia_link, '') FROM factions WHERE id = $1`, input.ID).
		Scan(&f.ID, &f.Name, &f.WahapediaLink)
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
		`INSERT INTO factions (id, name, wahapedia_link) VALUES ($1, $2, NULLIF($3, ''))`,
		f.ID, f.Name, f.WahapediaLink)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	return &FactionOutput{Body: f}, nil
}

func (h *AdminHandler) UpdateFaction(ctx context.Context, input *IDFactionInput) (*FactionOutput, error) {
	f := input.Body
	tag, err := h.db.Exec(ctx,
		`UPDATE factions SET name = $1, wahapedia_link = NULLIF($2, '') WHERE id = $3`,
		f.Name, f.WahapediaLink, input.ID)
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
	var query string
	var args []any
	if input.FactionID != "" {
		query = `SELECT id, faction_id, name FROM detachments WHERE faction_id = $1 ORDER BY name`
		args = []any{input.FactionID}
	} else {
		query = `SELECT id, faction_id, name FROM detachments ORDER BY faction_id, name`
	}
	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	detachments := make([]models.Detachment, 0)
	for rows.Next() {
		var d models.Detachment
		if err := rows.Scan(&d.ID, &d.FactionID, &d.Name); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		detachments = append(detachments, d)
	}
	return &DetachmentListOutput{Body: detachments}, nil
}

func (h *AdminHandler) GetDetachment(ctx context.Context, input *IDParam) (*DetachmentOutput, error) {
	var d models.Detachment
	err := h.db.QueryRow(ctx, `SELECT id, faction_id, name FROM detachments WHERE id = $1`, input.ID).
		Scan(&d.ID, &d.FactionID, &d.Name)
	if err != nil {
		return nil, huma.Error404NotFound("not found")
	}
	return &DetachmentOutput{Body: d}, nil
}

func (h *AdminHandler) CreateDetachment(ctx context.Context, input *DetachmentInput) (*DetachmentOutput, error) {
	d := input.Body
	if d.ID == "" || d.FactionID == "" || d.Name == "" {
		return nil, huma.Error400BadRequest("id, factionId, and name are required")
	}
	_, err := h.db.Exec(ctx,
		`INSERT INTO detachments (id, faction_id, name) VALUES ($1, $2, $3)`,
		d.ID, d.FactionID, d.Name)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	return &DetachmentOutput{Body: d}, nil
}

func (h *AdminHandler) UpdateDetachment(ctx context.Context, input *IDDetachmentInput) (*DetachmentOutput, error) {
	d := input.Body
	tag, err := h.db.Exec(ctx,
		`UPDATE detachments SET faction_id = $1, name = $2 WHERE id = $3`,
		d.FactionID, d.Name, input.ID)
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

// --- Stratagems ---

func (h *AdminHandler) ListStratagems(ctx context.Context, input *AdminStratagemListInput) (*StratagemListOutput, error) {
	query := `SELECT id, COALESCE(faction_id, ''), COALESCE(detachment_id, ''), name, type, cp_cost, COALESCE(legend, ''), turn, phase, description FROM stratagems`
	var args []any
	var conditions []string

	if input.FactionID != "" {
		args = append(args, input.FactionID)
		conditions = append(conditions, fmt.Sprintf("faction_id = $%d", len(args)))
	}
	if input.DetachmentID != "" {
		args = append(args, input.DetachmentID)
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

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	stratagems := make([]models.Stratagem, 0)
	for rows.Next() {
		var s models.Stratagem
		if err := rows.Scan(&s.ID, &s.FactionID, &s.DetachmentID, &s.Name, &s.Type, &s.CPCost, &s.Legend, &s.Turn, &s.Phase, &s.Description); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		stratagems = append(stratagems, s)
	}
	return &StratagemListOutput{Body: stratagems}, nil
}

func (h *AdminHandler) GetStratagem(ctx context.Context, input *IDParam) (*StratagemOutput, error) {
	var s models.Stratagem
	err := h.db.QueryRow(ctx,
		`SELECT id, COALESCE(faction_id, ''), COALESCE(detachment_id, ''), name, type, cp_cost, COALESCE(legend, ''), turn, phase, description FROM stratagems WHERE id = $1`, input.ID).
		Scan(&s.ID, &s.FactionID, &s.DetachmentID, &s.Name, &s.Type, &s.CPCost, &s.Legend, &s.Turn, &s.Phase, &s.Description)
	if err != nil {
		return nil, huma.Error404NotFound("not found")
	}
	return &StratagemOutput{Body: s}, nil
}

func (h *AdminHandler) CreateStratagem(ctx context.Context, input *StratagemInput) (*StratagemOutput, error) {
	s := input.Body
	if s.ID == "" || s.Name == "" {
		return nil, huma.Error400BadRequest("id and name are required")
	}
	_, err := h.db.Exec(ctx,
		`INSERT INTO stratagems (id, faction_id, detachment_id, name, type, cp_cost, legend, turn, phase, description)
		 VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), $4, $5, $6, NULLIF($7, ''), $8, $9, $10)`,
		s.ID, s.FactionID, s.DetachmentID, s.Name, s.Type, s.CPCost, s.Legend, s.Turn, s.Phase, s.Description)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	return &StratagemOutput{Body: s}, nil
}

func (h *AdminHandler) UpdateStratagem(ctx context.Context, input *IDStratagemInput) (*StratagemOutput, error) {
	s := input.Body
	tag, err := h.db.Exec(ctx,
		`UPDATE stratagems SET faction_id = NULLIF($1, ''), detachment_id = NULLIF($2, ''), name = $3, type = $4, cp_cost = $5, legend = NULLIF($6, ''), turn = $7, phase = $8, description = $9 WHERE id = $10`,
		s.FactionID, s.DetachmentID, s.Name, s.Type, s.CPCost, s.Legend, s.Turn, s.Phase, s.Description, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	s.ID = input.ID
	return &StratagemOutput{Body: s}, nil
}

func (h *AdminHandler) DeleteStratagem(ctx context.Context, input *IDParam) (*struct{}, error) {
	tag, err := h.db.Exec(ctx, `DELETE FROM stratagems WHERE id = $1`, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	return nil, nil
}

// --- Mission Packs ---

func (h *AdminHandler) ListMissionPacks(ctx context.Context, input *struct{}) (*MissionPackListOutput, error) {
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

func (h *AdminHandler) CreateMissionPack(ctx context.Context, input *MissionPackInput) (*MissionPackOutput, error) {
	p := input.Body
	if p.ID == "" || p.Name == "" {
		return nil, huma.Error400BadRequest("id and name are required")
	}
	_, err := h.db.Exec(ctx,
		`INSERT INTO mission_packs (id, name, description) VALUES ($1, $2, NULLIF($3, ''))`,
		p.ID, p.Name, p.Description)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	return &MissionPackOutput{Body: p}, nil
}

func (h *AdminHandler) UpdateMissionPack(ctx context.Context, input *IDMissionPackInput) (*MissionPackOutput, error) {
	p := input.Body
	tag, err := h.db.Exec(ctx,
		`UPDATE mission_packs SET name = $1, description = NULLIF($2, '') WHERE id = $3`,
		p.Name, p.Description, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	p.ID = input.ID
	return &MissionPackOutput{Body: p}, nil
}

func (h *AdminHandler) DeleteMissionPack(ctx context.Context, input *IDParam) (*struct{}, error) {
	tag, err := h.db.Exec(ctx, `DELETE FROM mission_packs WHERE id = $1`, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	return nil, nil
}

// --- Missions ---

func (h *AdminHandler) ListMissions(ctx context.Context, input *AdminPackFilterInput) (*MissionListOutput, error) {
	var query string
	var args []any
	if input.PackID != "" {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, ''), COALESCE(scoring_rules, '[]'), COALESCE(scoring_timing, 'end_of_command_phase') FROM missions WHERE mission_pack_id = $1 ORDER BY name`
		args = []any{input.PackID}
	} else {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, ''), COALESCE(scoring_rules, '[]'), COALESCE(scoring_timing, 'end_of_command_phase') FROM missions ORDER BY mission_pack_id, name`
	}
	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	missions := make([]models.Mission, 0)
	for rows.Next() {
		var m models.Mission
		var scoringRulesJSON string
		if err := rows.Scan(&m.ID, &m.MissionPackID, &m.Name, &m.Lore, &m.Description, &scoringRulesJSON, &m.ScoringTiming); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		json.Unmarshal([]byte(scoringRulesJSON), &m.ScoringRules)
		if m.ScoringRules == nil {
			m.ScoringRules = []models.ScoringAction{}
		}
		missions = append(missions, m)
	}
	return &MissionListOutput{Body: missions}, nil
}

func (h *AdminHandler) GetMission(ctx context.Context, input *IDParam) (*MissionOutput, error) {
	var m models.Mission
	var scoringRulesJSON string
	err := h.db.QueryRow(ctx,
		`SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, ''), COALESCE(scoring_rules, '[]'), COALESCE(scoring_timing, 'end_of_command_phase') FROM missions WHERE id = $1`, input.ID).
		Scan(&m.ID, &m.MissionPackID, &m.Name, &m.Lore, &m.Description, &scoringRulesJSON, &m.ScoringTiming)
	if err != nil {
		return nil, huma.Error404NotFound("not found")
	}
	json.Unmarshal([]byte(scoringRulesJSON), &m.ScoringRules)
	if m.ScoringRules == nil {
		m.ScoringRules = []models.ScoringAction{}
	}
	return &MissionOutput{Body: m}, nil
}

func (h *AdminHandler) CreateMission(ctx context.Context, input *MissionInput) (*MissionOutput, error) {
	m := input.Body
	if m.ID == "" || m.MissionPackID == "" || m.Name == "" {
		return nil, huma.Error400BadRequest("id, missionPackId, and name are required")
	}
	if m.ScoringRules == nil {
		m.ScoringRules = []models.ScoringAction{}
	}
	if m.ScoringTiming == "" {
		m.ScoringTiming = "end_of_command_phase"
	}
	scoringJSON, _ := json.Marshal(m.ScoringRules)
	_, err := h.db.Exec(ctx,
		`INSERT INTO missions (id, mission_pack_id, name, lore, description, scoring_rules, scoring_timing) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		m.ID, m.MissionPackID, m.Name, m.Lore, m.Description, string(scoringJSON), m.ScoringTiming)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	return &MissionOutput{Body: m}, nil
}

func (h *AdminHandler) UpdateMission(ctx context.Context, input *IDMissionInput) (*MissionOutput, error) {
	m := input.Body
	if m.ScoringRules == nil {
		m.ScoringRules = []models.ScoringAction{}
	}
	if m.ScoringTiming == "" {
		m.ScoringTiming = "end_of_command_phase"
	}
	scoringJSON, _ := json.Marshal(m.ScoringRules)
	tag, err := h.db.Exec(ctx,
		`UPDATE missions SET mission_pack_id = $1, name = $2, lore = $3, description = $4, scoring_rules = $5, scoring_timing = $6 WHERE id = $7`,
		m.MissionPackID, m.Name, m.Lore, m.Description, string(scoringJSON), m.ScoringTiming, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	m.ID = input.ID
	return &MissionOutput{Body: m}, nil
}

func (h *AdminHandler) DeleteMission(ctx context.Context, input *IDParam) (*struct{}, error) {
	tag, err := h.db.Exec(ctx, `DELETE FROM missions WHERE id = $1`, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	return nil, nil
}

// --- Secondaries ---

func (h *AdminHandler) ListSecondaries(ctx context.Context, input *AdminPackFilterInput) (*SecondaryListOutput, error) {
	var query string
	var args []any
	if input.PackID != "" {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, ''), max_vp, is_fixed, COALESCE(scoring_options, '[]') FROM secondaries WHERE mission_pack_id = $1 ORDER BY name`
		args = []any{input.PackID}
	} else {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, ''), max_vp, is_fixed, COALESCE(scoring_options, '[]') FROM secondaries ORDER BY mission_pack_id, name`
	}
	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	secondaries := make([]models.Secondary, 0)
	for rows.Next() {
		var s models.Secondary
		var scoringJSON string
		if err := rows.Scan(&s.ID, &s.MissionPackID, &s.Name, &s.Lore, &s.Description, &s.MaxVP, &s.IsFixed, &scoringJSON); err != nil {
			return nil, huma.Error500InternalServerError("scan error")
		}
		json.Unmarshal([]byte(scoringJSON), &s.ScoringOptions)
		if s.ScoringOptions == nil {
			s.ScoringOptions = []models.ScoringOption{}
		}
		secondaries = append(secondaries, s)
	}
	return &SecondaryListOutput{Body: secondaries}, nil
}

func (h *AdminHandler) GetSecondary(ctx context.Context, input *IDParam) (*SecondaryOutput, error) {
	var s models.Secondary
	var scoringJSON string
	err := h.db.QueryRow(ctx,
		`SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, ''), max_vp, is_fixed, COALESCE(scoring_options, '[]') FROM secondaries WHERE id = $1`, input.ID).
		Scan(&s.ID, &s.MissionPackID, &s.Name, &s.Lore, &s.Description, &s.MaxVP, &s.IsFixed, &scoringJSON)
	if err != nil {
		return nil, huma.Error404NotFound("not found")
	}
	json.Unmarshal([]byte(scoringJSON), &s.ScoringOptions)
	if s.ScoringOptions == nil {
		s.ScoringOptions = []models.ScoringOption{}
	}
	return &SecondaryOutput{Body: s}, nil
}

func (h *AdminHandler) CreateSecondary(ctx context.Context, input *SecondaryInput) (*SecondaryOutput, error) {
	s := input.Body
	if s.ID == "" || s.MissionPackID == "" || s.Name == "" {
		return nil, huma.Error400BadRequest("id, missionPackId, and name are required")
	}
	if s.ScoringOptions == nil {
		s.ScoringOptions = []models.ScoringOption{}
	}
	scoringJSON, _ := json.Marshal(s.ScoringOptions)
	_, err := h.db.Exec(ctx,
		`INSERT INTO secondaries (id, mission_pack_id, name, lore, description, max_vp, is_fixed, scoring_options) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		s.ID, s.MissionPackID, s.Name, s.Lore, s.Description, s.MaxVP, s.IsFixed, string(scoringJSON))
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	return &SecondaryOutput{Body: s}, nil
}

func (h *AdminHandler) UpdateSecondary(ctx context.Context, input *IDSecondaryInput) (*SecondaryOutput, error) {
	s := input.Body
	if s.ScoringOptions == nil {
		s.ScoringOptions = []models.ScoringOption{}
	}
	scoringJSON, _ := json.Marshal(s.ScoringOptions)
	tag, err := h.db.Exec(ctx,
		`UPDATE secondaries SET mission_pack_id = $1, name = $2, lore = $3, description = $4, max_vp = $5, is_fixed = $6, scoring_options = $7 WHERE id = $8`,
		s.MissionPackID, s.Name, s.Lore, s.Description, s.MaxVP, s.IsFixed, string(scoringJSON), input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	s.ID = input.ID
	return &SecondaryOutput{Body: s}, nil
}

func (h *AdminHandler) DeleteSecondary(ctx context.Context, input *IDParam) (*struct{}, error) {
	tag, err := h.db.Exec(ctx, `DELETE FROM secondaries WHERE id = $1`, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	return nil, nil
}

// --- Gambits ---

func (h *AdminHandler) ListGambits(ctx context.Context, input *AdminPackFilterInput) (*GambitListOutput, error) {
	var query string
	var args []any
	if input.PackID != "" {
		query = `SELECT id, mission_pack_id, name, COALESCE(description, ''), vp_value FROM gambits WHERE mission_pack_id = $1 ORDER BY name`
		args = []any{input.PackID}
	} else {
		query = `SELECT id, mission_pack_id, name, COALESCE(description, ''), vp_value FROM gambits ORDER BY mission_pack_id, name`
	}
	rows, err := h.db.Query(ctx, query, args...)
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

func (h *AdminHandler) GetGambit(ctx context.Context, input *IDParam) (*GambitOutput, error) {
	var g models.Gambit
	err := h.db.QueryRow(ctx,
		`SELECT id, mission_pack_id, name, COALESCE(description, ''), vp_value FROM gambits WHERE id = $1`, input.ID).
		Scan(&g.ID, &g.MissionPackID, &g.Name, &g.Description, &g.VPValue)
	if err != nil {
		return nil, huma.Error404NotFound("not found")
	}
	return &GambitOutput{Body: g}, nil
}

func (h *AdminHandler) CreateGambit(ctx context.Context, input *GambitInput) (*GambitOutput, error) {
	g := input.Body
	if g.ID == "" || g.MissionPackID == "" || g.Name == "" {
		return nil, huma.Error400BadRequest("id, missionPackId, and name are required")
	}
	_, err := h.db.Exec(ctx,
		`INSERT INTO gambits (id, mission_pack_id, name, description, vp_value) VALUES ($1, $2, $3, $4, $5)`,
		g.ID, g.MissionPackID, g.Name, g.Description, g.VPValue)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	return &GambitOutput{Body: g}, nil
}

func (h *AdminHandler) UpdateGambit(ctx context.Context, input *IDGambitInput) (*GambitOutput, error) {
	g := input.Body
	tag, err := h.db.Exec(ctx,
		`UPDATE gambits SET mission_pack_id = $1, name = $2, description = $3, vp_value = $4 WHERE id = $5`,
		g.MissionPackID, g.Name, g.Description, g.VPValue, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	g.ID = input.ID
	return &GambitOutput{Body: g}, nil
}

func (h *AdminHandler) DeleteGambit(ctx context.Context, input *IDParam) (*struct{}, error) {
	tag, err := h.db.Exec(ctx, `DELETE FROM gambits WHERE id = $1`, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	return nil, nil
}

// --- Challenger Cards ---

func (h *AdminHandler) ListChallengerCards(ctx context.Context, input *AdminPackFilterInput) (*ChallengerCardListOutput, error) {
	var query string
	var args []any
	if input.PackID != "" {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, '') FROM challenger_cards WHERE mission_pack_id = $1 ORDER BY name`
		args = []any{input.PackID}
	} else {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, '') FROM challenger_cards ORDER BY mission_pack_id, name`
	}
	rows, err := h.db.Query(ctx, query, args...)
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

func (h *AdminHandler) GetChallengerCard(ctx context.Context, input *IDParam) (*ChallengerCardOutput, error) {
	var c models.ChallengerCard
	err := h.db.QueryRow(ctx,
		`SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, '') FROM challenger_cards WHERE id = $1`, input.ID).
		Scan(&c.ID, &c.MissionPackID, &c.Name, &c.Lore, &c.Description)
	if err != nil {
		return nil, huma.Error404NotFound("not found")
	}
	return &ChallengerCardOutput{Body: c}, nil
}

func (h *AdminHandler) CreateChallengerCard(ctx context.Context, input *ChallengerCardInput) (*ChallengerCardOutput, error) {
	c := input.Body
	if c.ID == "" || c.MissionPackID == "" || c.Name == "" {
		return nil, huma.Error400BadRequest("id, missionPackId, and name are required")
	}
	_, err := h.db.Exec(ctx,
		`INSERT INTO challenger_cards (id, mission_pack_id, name, lore, description) VALUES ($1, $2, $3, $4, $5)`,
		c.ID, c.MissionPackID, c.Name, c.Lore, c.Description)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	return &ChallengerCardOutput{Body: c}, nil
}

func (h *AdminHandler) UpdateChallengerCard(ctx context.Context, input *IDChallengerCardInput) (*ChallengerCardOutput, error) {
	c := input.Body
	tag, err := h.db.Exec(ctx,
		`UPDATE challenger_cards SET mission_pack_id = $1, name = $2, lore = $3, description = $4 WHERE id = $5`,
		c.MissionPackID, c.Name, c.Lore, c.Description, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	c.ID = input.ID
	return &ChallengerCardOutput{Body: c}, nil
}

func (h *AdminHandler) DeleteChallengerCard(ctx context.Context, input *IDParam) (*struct{}, error) {
	tag, err := h.db.Exec(ctx, `DELETE FROM challenger_cards WHERE id = $1`, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	return nil, nil
}

// --- Mission Rules ---

func (h *AdminHandler) ListMissionRules(ctx context.Context, input *AdminPackFilterInput) (*MissionRuleListOutput, error) {
	var query string
	var args []any
	if input.PackID != "" {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, '') FROM mission_rules WHERE mission_pack_id = $1 ORDER BY name`
		args = []any{input.PackID}
	} else {
		query = `SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, '') FROM mission_rules ORDER BY mission_pack_id, name`
	}
	rows, err := h.db.Query(ctx, query, args...)
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

func (h *AdminHandler) GetMissionRule(ctx context.Context, input *IDParam) (*MissionRuleOutput, error) {
	var mr models.MissionRule
	err := h.db.QueryRow(ctx,
		`SELECT id, mission_pack_id, name, COALESCE(lore, ''), COALESCE(description, '') FROM mission_rules WHERE id = $1`, input.ID).
		Scan(&mr.ID, &mr.MissionPackID, &mr.Name, &mr.Lore, &mr.Description)
	if err != nil {
		return nil, huma.Error404NotFound("not found")
	}
	return &MissionRuleOutput{Body: mr}, nil
}

func (h *AdminHandler) CreateMissionRule(ctx context.Context, input *MissionRuleInput) (*MissionRuleOutput, error) {
	mr := input.Body
	if mr.ID == "" || mr.MissionPackID == "" || mr.Name == "" {
		return nil, huma.Error400BadRequest("id, missionPackId, and name are required")
	}
	_, err := h.db.Exec(ctx,
		`INSERT INTO mission_rules (id, mission_pack_id, name, lore, description) VALUES ($1, $2, $3, $4, $5)`,
		mr.ID, mr.MissionPackID, mr.Name, mr.Lore, mr.Description)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	return &MissionRuleOutput{Body: mr}, nil
}

func (h *AdminHandler) UpdateMissionRule(ctx context.Context, input *IDMissionRuleInput) (*MissionRuleOutput, error) {
	mr := input.Body
	tag, err := h.db.Exec(ctx,
		`UPDATE mission_rules SET mission_pack_id = $1, name = $2, lore = $3, description = $4 WHERE id = $5`,
		mr.MissionPackID, mr.Name, mr.Lore, mr.Description, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	mr.ID = input.ID
	return &MissionRuleOutput{Body: mr}, nil
}

func (h *AdminHandler) DeleteMissionRule(ctx context.Context, input *IDParam) (*struct{}, error) {
	tag, err := h.db.Exec(ctx, `DELETE FROM mission_rules WHERE id = $1`, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error: " + err.Error())
	}
	if tag.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("not found")
	}
	return nil, nil
}
