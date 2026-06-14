package game

import (
	"time"

	"github.com/peter/tacticarium/backend/internal/game/scoring"
)

type Phase string

const (
	PhaseSetup       Phase = "setup"
	PhaseStartOfTurn Phase = "start_of_turn"
	PhaseCommand     Phase = "command"
	PhaseMovement    Phase = "movement"
	PhaseShooting    Phase = "shooting"
	PhaseCharge      Phase = "charge"
	PhaseFight       Phase = "fight"
	PhaseEndOfTurn   Phase = "end_of_turn"
)

type GameStatus string

const (
	StatusSetup     GameStatus = "setup"
	StatusActive    GameStatus = "active"
	StatusCompleted GameStatus = "completed"
	StatusAbandoned GameStatus = "abandoned"
)

// SecondaryCard is a card in a player's secondary deck/hand/scored pile. It
// embeds the evaluable card (id, name, awards, text) and, once scored, the VP it
// yielded.
type SecondaryCard struct {
	scoring.Card
	VPScored int `json:"vpScored,omitempty"`
}

// ScorePrompt is an outstanding request for a player to confirm an off-board
// (Layer-2) fact so an award can be scored. The player responds with
// confirm_award supplying a count.
type ScorePrompt struct {
	ID         string `json:"id"`
	Category   string `json:"category"` // "primary" | "secondary"
	CardID     string `json:"cardId"`
	CardName   string `json:"cardName"`
	AwardIndex int    `json:"awardIndex"`
	Round      int    `json:"round"`
	Text       string `json:"text"`
}

// SelectedDetachment is one detachment a player has taken. 11e lets a player
// combine detachments up to a points budget (see MaxDetachmentPoints); Points is
// the detachment's cost (1–3).
type SelectedDetachment struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Points int    `json:"points"`
}

type PlayerState struct {
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	AvatarURL    string `json:"avatarUrl,omitempty"`
	PlayerNumber int    `json:"playerNumber"`
	FactionID    string `json:"factionId"`
	FactionName  string `json:"factionName"`
	// Detachments are the detachments this player has taken (combined up to
	// MaxDetachmentPoints). Their force dispositions union to form the player's
	// disposition choices.
	Detachments []SelectedDetachment `json:"detachments"`

	// 11e: each player has a side, a chosen force disposition, and a resolved
	// (asymmetric) primary mission.
	Side                 scoring.Side `json:"side,omitempty"`
	ForceDisposition     string       `json:"forceDisposition,omitempty"`
	ForceDispositionName string       `json:"forceDispositionName,omitempty"`
	MissionID            string       `json:"missionId,omitempty"`
	MissionName          string       `json:"missionName,omitempty"`
	// PrimaryCard is the scoring card for this player's resolved primary mission
	// (card_type="primary"); its awards drive primary VP.
	PrimaryCard scoring.Card `json:"primaryCard,omitempty"`

	CP          int  `json:"cp"`
	VPPrimary   int  `json:"vpPrimary"`
	VPSecondary int  `json:"vpSecondary"`
	VPPaint     int  `json:"vpPaint"`
	Ready       bool `json:"ready"`

	// Secondary deck (draw 2 per turn, keep unscored cards).
	SecondaryMode string `json:"secondaryMode"` // "fixed" | "tactical"
	// FixedSecondaryIDs are the card ids a fixed-mode player chose before the
	// game; at game start these are dealt to SecondaryHand and kept all game.
	// Empty for tactical players.
	FixedSecondaryIDs []string        `json:"fixedSecondaryIds"`
	SecondaryDeck     []SecondaryCard `json:"secondaryDeck"`
	SecondaryHand     []SecondaryCard `json:"secondaryHand"`
	SecondaryScored   []SecondaryCard `json:"secondaryScored"`

	CPGainedThisRound       int      `json:"cpGainedThisRound"`
	StratagemsUsedThisPhase []string `json:"stratagemsUsedThisPhase"`

	// Per-round VP accounting for the 15-VP/round caps (reset each battle round).
	// Per-game totals are VPPrimary / VPSecondary (capped at the game cap).
	PrimaryScoredThisRound   int `json:"primaryScoredThisRound"`
	SecondaryScoredThisRound int `json:"secondaryScoredThisRound"`

	// Outstanding Layer-2 scoring prompts awaiting player confirmation.
	PendingScorePrompts []ScorePrompt `json:"pendingScorePrompts"`
}

func (p *PlayerState) TotalVP() int {
	return p.VPPrimary + p.VPSecondary + p.VPPaint
}

type GameState struct {
	GameID          string        `json:"gameId"`
	InviteCode      string        `json:"inviteCode"`
	Status          GameStatus    `json:"status"`
	CurrentRound    int           `json:"currentRound"`
	CurrentTurn     int           `json:"currentTurn"`
	CurrentPhase    Phase         `json:"currentPhase"`
	ActivePlayer    int           `json:"activePlayer"`
	FirstTurnPlayer int           `json:"firstTurnPlayer"`
	Board           scoring.Board `json:"board"`
	// VP caps for this game, sourced from the mission cards (default 45 / 15).
	VPPerGameCap  int `json:"vpPerGameCap"`
	VPPerRoundCap int `json:"vpPerRoundCap"`
	// StartOfTurnControl snapshots objective control (index -> player) at the
	// start of the current turn, for "newly controlled this turn" scoring.
	StartOfTurnControl map[int]int `json:"startOfTurnControl,omitempty"`

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

// gameCap / roundCap return the configured caps, falling back to defaults.
func (gs *GameState) gameCap() int {
	if gs.VPPerGameCap > 0 {
		return gs.VPPerGameCap
	}
	return DefaultVPPerGameCap
}

func (gs *GameState) roundCap() int {
	if gs.VPPerRoundCap > 0 {
		return gs.VPPerRoundCap
	}
	return DefaultVPPerRoundCap
}
