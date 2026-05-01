package models

import "time"

type User struct {
	ID              string    `json:"id"`
	DiscordID       string    `json:"discordId"`
	DiscordUsername string    `json:"discordUsername"`
	DiscordAvatar   *string   `json:"discordAvatar,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Faction struct {
	ID            string `json:"id" required:"false"`
	Name          string `json:"name"`
	WahapediaLink string `json:"wahapediaLink,omitempty"`
}

type Detachment struct {
	ID        string `json:"id" required:"false"`
	FactionID string `json:"factionId"`
	Name      string `json:"name"`
	GameMode  string `json:"gameMode,omitempty"`
}

type Stratagem struct {
	ID           string `json:"id" required:"false"`
	FactionID    string `json:"factionId"`
	DetachmentID string `json:"detachmentId,omitempty"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	CPCost       int    `json:"cpCost"`
	Legend       string `json:"legend,omitempty"`
	Turn         string `json:"turn"`
	Phase        string `json:"phase"`
	Description  string `json:"description"`
	GameMode     string `json:"gameMode,omitempty"`
}

type MissionPack struct {
	ID          string `json:"id" required:"false"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ScoringAction struct {
	Label         string `json:"label"`
	VP            int    `json:"vp"`
	MinRound      int    `json:"minRound,omitempty"`
	Desc          string `json:"description,omitempty"`
	ScoringTiming string `json:"scoringTiming,omitempty"`
}

type Mission struct {
	ID            string          `json:"id" required:"false"`
	MissionPackID string          `json:"missionPackId"`
	Name          string          `json:"name"`
	Lore          string          `json:"lore,omitempty"`
	Description   string          `json:"description"`
	ScoringRules  []ScoringAction `json:"scoringRules"`
	ScoringTiming string          `json:"scoringTiming"`
}

type MissionRule struct {
	ID            string `json:"id" required:"false"`
	MissionPackID string `json:"missionPackId"`
	Name          string `json:"name"`
	Lore          string `json:"lore,omitempty"`
	Description   string `json:"description"`
}

type ScoringOption struct {
	Label string `json:"label"`
	VP    int    `json:"vp"`
	Mode  string `json:"mode,omitempty"` // "fixed", "tactical", or "" (both)
}

type DrawRestriction struct {
	Round int    `json:"round"`
	Mode  string `json:"mode"` // "mandatory" or "optional"
}

type Secondary struct {
	ID              string           `json:"id" required:"false"`
	MissionPackID   string           `json:"missionPackId"`
	Name            string           `json:"name"`
	Lore            string           `json:"lore,omitempty"`
	Description     string           `json:"description"`
	MaxVP           int              `json:"maxVp"`
	IsFixed         bool             `json:"isFixed"`
	ScoringOptions  []ScoringOption  `json:"scoringOptions"`
	DrawRestriction *DrawRestriction `json:"drawRestriction,omitempty"`
	// ScoringTiming controls when the frontend prompts the owner to score.
	// "end_of_own_turn" (default) or "end_of_opponent_turn".
	ScoringTiming string `json:"scoringTiming,omitempty"`
}

type ChallengerCard struct {
	ID            string `json:"id" required:"false"`
	MissionPackID string `json:"missionPackId"`
	Name          string `json:"name"`
	Lore          string `json:"lore,omitempty"`
	Description   string `json:"description"`
}

type Gambit struct {
	ID            string `json:"id" required:"false"`
	MissionPackID string `json:"missionPackId"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	VPValue       int    `json:"vpValue"`
}

type GameSummary struct {
	ID          string              `json:"id"`
	InviteCode  string              `json:"inviteCode"`
	Status      string              `json:"status"`
	MissionName string              `json:"missionName,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
	CompletedAt *time.Time          `json:"completedAt,omitempty"`
	Players     []GamePlayerSummary `json:"players"`
	WinnerID    *string             `json:"winnerId,omitempty"`
}

type GamePlayerSummary struct {
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	FactionName  string `json:"factionName,omitempty"`
	PlayerNumber int    `json:"playerNumber"`
	TotalVP      int    `json:"totalVp"`
}
