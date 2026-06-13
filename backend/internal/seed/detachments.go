package seed

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type detachmentJSON struct {
	ID                string      `json:"id"`
	FactionID         string      `json:"faction_id"`
	Name              string      `json:"name"`
	DetachmentPoints  *int        `json:"detachment_points"`
	ForceDispositions []string    `json:"force_dispositions"`
	GameVersion       gameVersion `json:"game_version"`
}

// SeedDetachments upserts every detachment in a 40kdc-data detachments.json
// array. Factions must be seeded first (faction_id is a FK).
func SeedDetachments(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	var detachments []detachmentJSON
	if err := readJSON(filePath, &detachments); err != nil {
		return 0, err
	}

	count := 0
	for _, d := range detachments {
		if d.ID == "" || d.FactionID == "" {
			continue
		}
		dispositions := d.ForceDispositions
		if dispositions == nil {
			dispositions = []string{}
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO detachments (id, faction_id, name, detachment_points, force_dispositions)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (id) DO UPDATE
			   SET faction_id = $2, name = $3, detachment_points = $4, force_dispositions = $5`,
			d.ID, d.FactionID, d.Name, d.DetachmentPoints, dispositions)
		if err != nil {
			return count, fmt.Errorf("inserting detachment %s: %w", d.ID, err)
		}
		count++
	}
	return count, nil
}
