package seed

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedStratagems reads Stratagems.csv and upserts every stratagem. Detachments
// must already be seeded (see SeedDetachments) — the stratagem seed looks up
// each detachment's game_mode from the DB to decide the stratagem's game_mode.
//
// A stratagem is tagged boarding_actions if either:
//   - its type starts with "Boarding Actions – " (the 6 generic BA strats), or
//   - its detachment_id points to a boarding_actions detachment.
//
// Otherwise it is tagged core.
func SeedStratagems(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("opening stratagems file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = '|'
	reader.LazyQuotes = true

	// Skip header
	if _, err := reader.Read(); err != nil {
		return 0, fmt.Errorf("reading header: %w", err)
	}

	// Load detachment game modes from the DB so we can tag stratagems without
	// a round-trip per row. SeedDetachments must have run first for this to be
	// non-empty; if it hasn't, stratagems attached to BA-only detachments will
	// default to core. We warn rather than fail so seeding stratagems
	// independently still works in ad-hoc scenarios.
	detachmentModes, err := loadDetachmentModesFromDB(ctx, pool)
	if err != nil {
		return 0, fmt.Errorf("loading detachment game modes: %w", err)
	}

	stratagemCount := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if len(record) < 11 {
			continue
		}

		factionID := strings.TrimSpace(record[0])
		name := strings.TrimSpace(record[1])
		id := strings.TrimSpace(record[2])
		stratType := strings.TrimSpace(record[3])
		cpCostStr := strings.TrimSpace(record[4])
		legend := strings.TrimSpace(record[5])
		turn := strings.ReplaceAll(strings.TrimSpace(record[6]), "\u2019", "'")
		phase := strings.TrimSpace(record[7])
		detachmentID := strings.TrimSpace(record[9])
		description := strings.TrimSpace(record[10])

		if id == "" || name == "" {
			continue
		}

		cpCost, _ := strconv.Atoi(cpCostStr)

		var factionIDPtr, detachmentIDPtr *string
		if factionID != "" {
			factionIDPtr = &factionID
		}
		if detachmentID != "" {
			detachmentIDPtr = &detachmentID
		}

		gameMode := "core"
		if strings.HasPrefix(stratType, "Boarding Actions \u2013 ") {
			gameMode = "boarding_actions"
		} else if mode, ok := detachmentModes[detachmentID]; ok && mode == "boarding_actions" {
			gameMode = "boarding_actions"
		}

		_, err = pool.Exec(ctx,
			`INSERT INTO stratagems (id, faction_id, detachment_id, name, type, cp_cost, legend, turn, phase, description, game_mode)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			 ON CONFLICT (id) DO UPDATE SET
			   faction_id = $2, detachment_id = $3, name = $4, type = $5,
			   cp_cost = $6, legend = $7, turn = $8, phase = $9, description = $10, game_mode = $11`,
			id, factionIDPtr, detachmentIDPtr, name, stratType, cpCost, legend, turn, phase, description, gameMode)
		if err != nil {
			fmt.Printf("Warning: error inserting stratagem %s: %v\n", id, err)
			continue
		}
		stratagemCount++
	}

	return stratagemCount, nil
}

func loadDetachmentModesFromDB(ctx context.Context, pool *pgxpool.Pool) (map[string]string, error) {
	rows, err := pool.Query(ctx, `SELECT id, game_mode FROM detachments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	modes := make(map[string]string)
	for rows.Next() {
		var id, mode string
		if err := rows.Scan(&id, &mode); err != nil {
			return nil, err
		}
		modes[id] = mode
	}
	return modes, rows.Err()
}
