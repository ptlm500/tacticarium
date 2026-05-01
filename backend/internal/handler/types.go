package handler

import (
	"time"

	"github.com/peter/tacticarium/backend/internal/game"
	"github.com/peter/tacticarium/backend/internal/models"
)

// --- Common ---

type IDParam struct {
	ID string `path:"id" doc:"Resource ID"`
}

// --- Health ---

type HealthOutput struct {
	Body struct {
		Status string `json:"status" example:"ok"`
	}
}

// --- Auth ---

type MeOutput struct {
	Body struct {
		ID        string    `json:"id"`
		Username  string    `json:"username"`
		Avatar    *string   `json:"avatar,omitempty"`
		CreatedAt time.Time `json:"createdAt"`
	}
}

type AdminMeOutput struct {
	Body struct {
		GitHubID   string `json:"githubId"`
		GitHubUser string `json:"githubUser"`
	}
}

// --- Factions ---

type FactionIDParam struct {
	FactionID string `path:"factionId" doc:"Faction ID"`
}

type DetachmentIDParam struct {
	DetachmentID string `path:"detachmentId" doc:"Detachment ID"`
}

type FactionListOutput struct {
	Body []models.Faction
}

type DetachmentListOutput struct {
	Body []models.Detachment
}

type StratagemListOutput struct {
	Body []models.Stratagem
}

// --- Missions ---

type PackIDParam struct {
	PackID string `path:"packId" doc:"Mission Pack ID"`
}

type MissionPackListOutput struct {
	Body []models.MissionPack
}

type MissionListOutput struct {
	Body []models.Mission
}

type SecondaryListOutput struct {
	Body []models.Secondary
}

type GambitListOutput struct {
	Body []models.Gambit
}

type MissionRuleListOutput struct {
	Body []models.MissionRule
}

type ChallengerCardListOutput struct {
	Body []models.ChallengerCard
}

// --- Games ---

type CreateGameOutput struct {
	Body struct {
		ID         string `json:"id"`
		InviteCode string `json:"inviteCode"`
	}
}

type JoinGameInput struct {
	Code string `path:"code" doc:"Invite code"`
}

type JoinGameOutput struct {
	Body struct {
		ID         string `json:"id"`
		InviteCode string `json:"inviteCode"`
	}
}

type GameIDParam struct {
	GameID string `path:"gameId" doc:"Game ID"`
}

type GameStateOutput struct {
	Body *game.GameState
}

type GameListOutput struct {
	Body []models.GameSummary
}

type GameEvent struct {
	ID           int64     `json:"id"`
	PlayerNumber *int      `json:"playerNumber"`
	EventType    string    `json:"eventType"`
	EventData    any       `json:"eventData"`
	Round        *int      `json:"round"`
	Phase        *string   `json:"phase"`
	CreatedAt    time.Time `json:"createdAt"`
}

type GameEventsOutput struct {
	Body []GameEvent
}

type HistoryInput struct {
	MyFaction       string `query:"myFaction" doc:"Filter by player's faction name"`
	OpponentFaction string `query:"opponentFaction" doc:"Filter by opponent's faction name"`
}

type FactionStat struct {
	FactionName string `json:"factionName"`
	GamesPlayed int    `json:"gamesPlayed"`
	Wins        int    `json:"wins"`
}

type UserStats struct {
	Wins         int           `json:"wins"`
	Losses       int           `json:"losses"`
	Draws        int           `json:"draws"`
	Abandoned    int           `json:"abandoned"`
	FactionStats []FactionStat `json:"factionStats"`
	AverageVP    float64       `json:"averageVp"`
}

type StatsOutput struct {
	Body UserStats
}

// --- Admin CRUD ---

// Inputs with body
type FactionInput struct {
	Body models.Faction
}

type DetachmentInput struct {
	Body models.Detachment
}

type StratagemInput struct {
	Body models.Stratagem
}

type MissionPackInput struct {
	Body models.MissionPack
}

type MissionInput struct {
	Body models.Mission
}

type SecondaryInput struct {
	Body models.Secondary
}

type GambitInput struct {
	Body models.Gambit
}

type ChallengerCardInput struct {
	Body models.ChallengerCard
}

type MissionRuleInput struct {
	Body models.MissionRule
}

// Inputs with path + body
type IDFactionInput struct {
	ID   string `path:"id" doc:"Resource ID"`
	Body models.Faction
}

type IDDetachmentInput struct {
	ID   string `path:"id" doc:"Resource ID"`
	Body models.Detachment
}

type IDStratagemInput struct {
	ID   string `path:"id" doc:"Resource ID"`
	Body models.Stratagem
}

type IDMissionPackInput struct {
	ID   string `path:"id" doc:"Resource ID"`
	Body models.MissionPack
}

type IDMissionInput struct {
	ID   string `path:"id" doc:"Resource ID"`
	Body models.Mission
}

type IDSecondaryInput struct {
	ID   string `path:"id" doc:"Resource ID"`
	Body models.Secondary
}

type IDGambitInput struct {
	ID   string `path:"id" doc:"Resource ID"`
	Body models.Gambit
}

type IDChallengerCardInput struct {
	ID   string `path:"id" doc:"Resource ID"`
	Body models.ChallengerCard
}

type IDMissionRuleInput struct {
	ID   string `path:"id" doc:"Resource ID"`
	Body models.MissionRule
}

// Single-item outputs
type FactionOutput struct {
	Body models.Faction
}

type DetachmentOutput struct {
	Body models.Detachment
}

type StratagemOutput struct {
	Body models.Stratagem
}

type MissionPackOutput struct {
	Body models.MissionPack
}

type MissionOutput struct {
	Body models.Mission
}

type SecondaryOutput struct {
	Body models.Secondary
}

type GambitOutput struct {
	Body models.Gambit
}

type ChallengerCardOutput struct {
	Body models.ChallengerCard
}

type MissionRuleOutput struct {
	Body models.MissionRule
}

// Admin list inputs with optional query filters
type AdminDetachmentListInput struct {
	FactionID string `query:"faction_id" doc:"Filter by faction ID"`
}

type AdminStratagemListInput struct {
	FactionID    string `query:"faction_id" doc:"Filter by faction ID"`
	DetachmentID string `query:"detachment_id" doc:"Filter by detachment ID"`
}

type AdminPackFilterInput struct {
	PackID string `query:"pack_id" doc:"Filter by mission pack ID"`
}

// Import outputs
type ImportResultOutput struct {
	Body map[string]any
}
