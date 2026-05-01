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

// SecondaryScoringOption represents one scoring criterion for a secondary mission.
type SecondaryScoringOption struct {
	Label string `json:"label"`
	VP    int    `json:"vp"`
	Mode  string `json:"mode,omitempty"` // "fixed", "tactical", or "" (both)
}

// DrawRestrictionMode enumerates how a draw restriction is enforced.
const (
	DrawRestrictionMandatory = "mandatory"
	DrawRestrictionOptional  = "optional"
)

// SecondaryDrawRestriction describes a "When Drawn" rule that triggers when a
// card is drawn during a specific battle round. Mandatory restrictions force
// the card to be shuffled back into the deck; optional restrictions let the
// player choose to shuffle it back via the reshuffle_secondary action.
type SecondaryDrawRestriction struct {
	Round int    `json:"round"`
	Mode  string `json:"mode"`
}

// ActiveSecondary represents a secondary mission card in the deck/active/achieved/discarded piles.
type ActiveSecondary struct {
	ID              string                    `json:"id"`
	Name            string                    `json:"name"`
	Description     string                    `json:"description"`
	IsFixed         bool                      `json:"isFixed"`
	MaxVP           int                       `json:"maxVp"`
	ScoringOptions  []SecondaryScoringOption  `json:"scoringOptions"`
	DrawRestriction *SecondaryDrawRestriction `json:"drawRestriction,omitempty"`
	// ScoringTiming controls when the frontend prompts the owner to score.
	// "end_of_own_turn" (default) or "end_of_opponent_turn".
	ScoringTiming string `json:"scoringTiming,omitempty"`
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

	// Mission system fields
	SecondaryMode           string            `json:"secondaryMode"`
	TacticalDeck            []ActiveSecondary `json:"tacticalDeck"`
	ActiveSecondaries       []ActiveSecondary `json:"activeSecondaries"`
	AchievedSecondaries     []ActiveSecondary `json:"achievedSecondaries"`
	DiscardedSecondaries    []ActiveSecondary `json:"discardedSecondaries"`
	CPGainedThisRound       int               `json:"cpGainedThisRound"`
	IsChallenger            bool              `json:"isChallenger"`
	ChallengerCardID        string            `json:"challengerCardId,omitempty"`
	AdaptOrDieUses          int               `json:"adaptOrDieUses"`
	StratagemsUsedThisPhase []string          `json:"stratagemsUsedThisPhase"`

	// VPPrimaryScoredSlots maps battle round -> scoring slot -> applied VP delta.
	// Used to prevent double-scoring the same primary slot in a round and to
	// support undoing a specific prior score.
	VPPrimaryScoredSlots map[int]map[string]int `json:"vpPrimaryScoredSlots"`
}

func (p *PlayerState) TotalVP() int {
	return p.VPPrimary + p.VPSecondary + p.VPGambit + p.VPPaint
}

type GameState struct {
	GameID             string          `json:"gameId"`
	InviteCode         string          `json:"inviteCode"`
	Status             GameStatus      `json:"status"`
	CurrentRound       int             `json:"currentRound"`
	CurrentTurn        int             `json:"currentTurn"`
	CurrentPhase       Phase           `json:"currentPhase"`
	ActivePlayer       int             `json:"activePlayer"`
	FirstTurnPlayer    int             `json:"firstTurnPlayer"`
	MissionPackID      string          `json:"missionPackId"`
	MissionID          string          `json:"missionId"`
	MissionName        string          `json:"missionName"`
	TwistID            string          `json:"twistId"`
	TwistName          string          `json:"twistName"`
	Players            [2]*PlayerState `json:"players"`
	CreatedAt          time.Time       `json:"createdAt"`
	CompletedAt        *time.Time      `json:"completedAt,omitempty"`
	WinnerID           string          `json:"winnerId,omitempty"`
	AbandonRequestedBy *int            `json:"abandonRequestedBy,omitempty"`
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
