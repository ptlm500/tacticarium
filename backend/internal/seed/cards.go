package seed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type cardJSON struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	CardType    string          `json:"card_type"`
	Subtype     *string         `json:"subtype"`
	WhenDrawn   json.RawMessage `json:"when_drawn"`
	Actions     json.RawMessage `json:"actions"`
	Awards      json.RawMessage `json:"awards"`
	Text        string          `json:"text"`
	GameVersion gameVersion     `json:"game_version"`
}

// SeedCards upserts every mission card (primary + secondary) from a 40kdc-data
// secondary-cards.json array. The awards / when_drawn / actions DSL blocks are
// stored as JSONB for the scoring evaluator.
func SeedCards(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	var cards []cardJSON
	if err := readJSON(filePath, &cards); err != nil {
		return 0, err
	}

	count := 0
	for _, c := range cards {
		if c.ID == "" {
			continue
		}
		cardType := c.CardType
		if cardType == "" {
			cardType = "secondary"
		}
		edition := c.GameVersion.Edition
		if edition == "" {
			edition = "11th"
		}
		dataslate := c.GameVersion.Dataslate
		if dataslate == "" {
			dataslate = "launch"
		}

		whenDrawn := rawOrNil(c.WhenDrawn)
		actions := rawOrDefault(c.Actions, "[]")
		awards := rawOrDefault(c.Awards, "[]")

		_, err := pool.Exec(ctx,
			`INSERT INTO cards (id, name, card_type, subtype, when_drawn, actions, awards, text, edition, dataslate)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 ON CONFLICT (id) DO UPDATE SET
			   name = $2, card_type = $3, subtype = $4, when_drawn = $5, actions = $6,
			   awards = $7, text = $8, edition = $9, dataslate = $10`,
			c.ID, c.Name, cardType, c.Subtype, whenDrawn, actions, awards, c.Text, edition, dataslate)
		if err != nil {
			return count, fmt.Errorf("inserting card %s: %w", c.ID, err)
		}
		count++
	}
	return count, nil
}

// rawOrNil returns the JSON text, or nil for an absent/null value (nullable
// jsonb columns).
func rawOrNil(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	return string(raw)
}

// rawOrDefault returns the JSON text, or def for an absent/null value
// (non-null jsonb columns with a default).
func rawOrDefault(raw json.RawMessage, def string) string {
	if len(raw) == 0 || string(raw) == "null" {
		return def
	}
	return string(raw)
}
