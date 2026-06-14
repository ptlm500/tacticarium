package scoring

import "encoding/json"

// Card is the subset of a 40kdc-data mission/secondary card the evaluator needs.
type Card struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	CardType string  `json:"card_type"`
	Awards   []Award `json:"awards"`
	Text     string  `json:"text"`
}

// Award is one VP-award block on a card.
type Award struct {
	Trigger        Trigger    `json:"trigger"`
	When           *Condition `json:"when,omitempty"`
	VP             *int       `json:"vp,omitempty"`
	VPPer          *int       `json:"vp_per,omitempty"`
	Per            string     `json:"per,omitempty"`
	PerMax         *int       `json:"per_max,omitempty"`
	VPMax          *int       `json:"vp_max,omitempty"`
	Mode           string     `json:"mode,omitempty"`
	Cumulative     bool       `json:"cumulative,omitempty"`
	ExclusiveGroup string     `json:"exclusive_group,omitempty"`
}

// Trigger is when an award is evaluated.
type Trigger struct {
	Phase       string       `json:"phase,omitempty"`
	Timing      string       `json:"timing,omitempty"`
	PlayerTurn  string       `json:"player_turn,omitempty"`
	BattleRound *RoundWindow `json:"battle_round,omitempty"`
}

// RoundWindow bounds the battle rounds in which a trigger/award is active.
type RoundWindow struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

// Condition is a `when` predicate node: either a leaf (Type + Parameters) or a
// compound (Operator + Operands).
type Condition struct {
	Type       string         `json:"type,omitempty"`
	Parameters map[string]any `json:"parameters,omitempty"`
	Negated    bool           `json:"negated,omitempty"`
	Operator   string         `json:"operator,omitempty"`
	Operands   []Condition    `json:"operands,omitempty"`
}

// ParseCard unmarshals a card from its JSON form. awards is the cards.awards
// JSONB column (or the secondary-cards.json `awards` array).
func ParseCard(id, name, cardType, text string, awards json.RawMessage) (Card, error) {
	c := Card{ID: id, Name: name, CardType: cardType, Text: text}
	if len(awards) > 0 && string(awards) != "null" {
		if err := json.Unmarshal(awards, &c.Awards); err != nil {
			return c, err
		}
	}
	return c, nil
}

// intParam reads an integer parameter (JSON numbers decode to float64).
func (c *Condition) intParam(key string) (int, bool) {
	v, ok := c.Parameters[key]
	if !ok {
		return 0, false
	}
	if f, ok := v.(float64); ok {
		return int(f), true
	}
	return 0, false
}

func (c *Condition) strParam(key string) (string, bool) {
	v, ok := c.Parameters[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
