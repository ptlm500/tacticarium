package game

import "time"

type Phase string

const (
	PhaseSetup    Phase = "setup"
	PhaseCommand  Phase = "command"
	PhaseMovement Phase = "movement"
	PhaseShooting Phase = "shooting"
	PhaseCharge   Phase = "charge"
	PhaseFight    Phase = "fight"
)

var PhaseOrder = []Phase{
	PhaseCommand, PhaseMovement, PhaseShooting, PhaseCharge, PhaseFight,
}

type GameStatus string

const (
	StatusSetup     GameStatus = "setup"
	StatusActive    GameStatus = "active"
	StatusCompleted GameStatus = "completed"
	StatusAbandoned GameStatus = "abandoned"
)

type SecondaryObjective struct {
	ID          string `json:"id"`
	SecondaryID string `json:"secondaryId,omitempty"`
	CustomName  string `json:"customName,omitempty"`
	CustomMaxVP int    `json:"customMaxVp,omitempty"`
	VPScored    int    `json:"vpScored"`
}

type PlayerState struct {
	UserID              string               `json:"userId"`
	Username            string               `json:"username"`
	PlayerNumber        int                  `json:"playerNumber"`
	FactionID           string               `json:"factionId"`
	FactionName         string               `json:"factionName"`
	DetachmentID        string               `json:"detachmentId"`
	DetachmentName      string               `json:"detachmentName"`
	CP                  int                  `json:"cp"`
	VPPrimary           int                  `json:"vpPrimary"`
	VPSecondary         int                  `json:"vpSecondary"`
	VPGambit            int                  `json:"vpGambit"`
	VPPaint             int                  `json:"vpPaint"`
	Ready               bool                 `json:"ready"`
	GambitID            string               `json:"gambitId,omitempty"`
	GambitDeclaredRound int                  `json:"gambitDeclaredRound,omitempty"`
	Secondaries         []SecondaryObjective `json:"secondaries"`
}

func (p *PlayerState) TotalVP() int {
	return p.VPPrimary + p.VPSecondary + p.VPGambit + p.VPPaint
}

type GameState struct {
	GameID          string          `json:"gameId"`
	InviteCode      string          `json:"inviteCode"`
	Status          GameStatus      `json:"status"`
	CurrentRound    int             `json:"currentRound"`
	CurrentPhase    Phase           `json:"currentPhase"`
	ActivePlayer    int             `json:"activePlayer"`
	FirstTurnPlayer int             `json:"firstTurnPlayer"`
	MissionPackID   string          `json:"missionPackId"`
	MissionID       string          `json:"missionId"`
	MissionName     string          `json:"missionName"`
	Players         [2]*PlayerState `json:"players"`
	CreatedAt       time.Time       `json:"createdAt"`
	CompletedAt     *time.Time      `json:"completedAt,omitempty"`
	WinnerID        string          `json:"winnerId,omitempty"`
}

func (gs *GameState) GetPlayer(playerNumber int) *PlayerState {
	if playerNumber == 1 || playerNumber == 2 {
		return gs.Players[playerNumber-1]
	}
	return nil
}

func (gs *GameState) GetPlayerByUserID(userID string) *PlayerState {
	for _, p := range gs.Players {
		if p != nil && p.UserID == userID {
			return p
		}
	}
	return nil
}
