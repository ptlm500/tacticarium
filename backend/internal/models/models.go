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
	ID            string `json:"id"`
	Name          string `json:"name"`
	WahapediaLink string `json:"wahapediaLink,omitempty"`
}

type Detachment struct {
	ID        string `json:"id"`
	FactionID string `json:"factionId"`
	Name      string `json:"name"`
}

type Stratagem struct {
	ID           string `json:"id"`
	FactionID    string `json:"factionId"`
	DetachmentID string `json:"detachmentId,omitempty"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	CPCost       int    `json:"cpCost"`
	Legend       string `json:"legend,omitempty"`
	Turn         string `json:"turn"`
	Phase        string `json:"phase"`
	Description  string `json:"description"`
}

type MissionPack struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Mission struct {
	ID            string `json:"id"`
	MissionPackID string `json:"missionPackId"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	DeploymentMap string `json:"deploymentMap,omitempty"`
	RulesText     string `json:"rulesText,omitempty"`
}

type Secondary struct {
	ID            string `json:"id"`
	MissionPackID string `json:"missionPackId"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	Description   string `json:"description"`
	MaxVP         int    `json:"maxVp"`
}

type Gambit struct {
	ID            string `json:"id"`
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
