package models

import (
	"encoding/json"
	"time"
)

type User struct {
	ID              string    `json:"id"`
	DiscordID       string    `json:"discordId"`
	DiscordUsername string    `json:"discordUsername"`
	DiscordAvatar   *string   `json:"discordAvatar,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Faction struct {
	ID              string `json:"id" required:"false"`
	Name            string `json:"name"`
	ParentFactionID string `json:"parentFactionId,omitempty"`
	FactionRuleID   string `json:"factionRuleId,omitempty"`
}

type Detachment struct {
	ID                string   `json:"id" required:"false"`
	FactionID         string   `json:"factionId"`
	Name              string   `json:"name"`
	DetachmentPoints  int      `json:"detachmentPoints,omitempty"`
	ForceDispositions []string `json:"forceDispositions"`
}

type Stratagem struct {
	ID           string   `json:"id" required:"false"`
	FactionID    string   `json:"factionId,omitempty"`
	DetachmentID string   `json:"detachmentId,omitempty"`
	Name         string   `json:"name"`
	Type         string   `json:"type,omitempty"`
	Category     string   `json:"category,omitempty"`
	CPCost       int      `json:"cpCost"`
	Phases       []string `json:"phases"`
	PlayerTurn   string   `json:"playerTurn,omitempty"`
	Timing       string   `json:"timing,omitempty"`
}

// ForceDisposition is one of the five 11e strategic-intent tags.
type ForceDisposition struct {
	ID   string `json:"id" required:"false"`
	Name string `json:"name"`
	Text string `json:"text,omitempty"`
}

// Mission is the 11e primary mission objective record.
type Mission struct {
	ID                   string   `json:"id" required:"false"`
	Name                 string   `json:"name"`
	VPPerGameCap         int      `json:"vpPerGameCap"`
	VPPerRoundCap        int      `json:"vpPerRoundCap"`
	DeploymentPatternIDs []string `json:"deploymentPatternIds"`
}

// MissionMatchup is one cell of the 5x5 disposition selector matrix.
type MissionMatchup struct {
	ID                  string `json:"id" required:"false"`
	Disposition         string `json:"disposition"`
	OpponentDisposition string `json:"opponentDisposition"`
	MissionID           string `json:"missionId"`
}

// Card is an 11e mission/secondary card (primary mission cards and the
// secondary deck). The awards DSL is internal to scoring; the API exposes the
// community-authored prose.
type MissionCard struct {
	ID       string `json:"id" required:"false"`
	Name     string `json:"name"`
	CardType string `json:"cardType"`
	Subtype  string `json:"subtype,omitempty"`
	Text     string `json:"text,omitempty"`
}

// DeploymentPattern carries the board geometry for the frontend board view.
type DeploymentPattern struct {
	ID                          string          `json:"id" required:"false"`
	Name                        string          `json:"name"`
	Source                      string          `json:"source,omitempty"`
	Description                 string          `json:"description,omitempty"`
	Objectives                  json.RawMessage `json:"objectives" doc:"Objective coordinates" type:"array"`
	Territories                 json.RawMessage `json:"territories" type:"array"`
	Zones                       json.RawMessage `json:"zones" type:"array"`
	RecommendedTerrainLayoutIDs []string        `json:"recommendedTerrainLayoutIds"`
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
	AvatarURL    string `json:"avatarUrl,omitempty"`
	FactionName  string `json:"factionName,omitempty"`
	PlayerNumber int    `json:"playerNumber"`
	TotalVP      int    `json:"totalVp"`
}
