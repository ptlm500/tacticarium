package seed

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedDetachments reads Detachments.csv and upserts every row. The CSV's
// `type` column is the authoritative source for game_mode: rows with
// type = "Boarding Actions" are tagged boarding_actions; everything else is
// core.
//
// Columns: id | faction_id | name | legend | type
func SeedDetachments(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("opening detachments file: %w", err)
	}
	defer func() { _ = f.Close() }()

	reader := csv.NewReader(f)
	reader.Comma = '|'
	reader.LazyQuotes = true
	// Rows have a trailing empty field from the terminating pipe; allow variable
	// field counts so csv doesn't complain on short/long rows.
	reader.FieldsPerRecord = -1

	// Skip header
	if _, err := reader.Read(); err != nil {
		return 0, fmt.Errorf("reading header: %w", err)
	}

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Skip malformed rows rather than aborting the whole seed.
			continue
		}

		if len(record) < 3 {
			continue
		}

		id := strings.TrimSpace(record[0])
		factionID := strings.TrimSpace(record[1])
		name := strings.TrimSpace(record[2])

		if id == "" || factionID == "" || name == "" {
			continue
		}

		gameMode := "core"
		if len(record) >= 5 && strings.TrimSpace(record[4]) == "Boarding Actions" {
			gameMode = "boarding_actions"
		}

		_, err = pool.Exec(ctx,
			`INSERT INTO detachments (id, faction_id, name, game_mode) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (id) DO UPDATE SET faction_id = $2, name = $3, game_mode = $4`,
			id, factionID, name, gameMode)
		if err != nil {
			fmt.Printf("Warning: error inserting detachment %s: %v\n", id, err)
			continue
		}
		count++
	}

	return count, nil
}
