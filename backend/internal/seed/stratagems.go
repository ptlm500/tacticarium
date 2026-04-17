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

func SeedStratagems(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, 0, fmt.Errorf("opening stratagems file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = '|'
	reader.LazyQuotes = true

	// Skip header
	if _, err := reader.Read(); err != nil {
		return 0, 0, fmt.Errorf("reading header: %w", err)
	}

	// First pass: extract detachments
	type detachmentInfo struct {
		ID        string
		FactionID string
		Name      string
	}

	detachmentMap := make(map[string]detachmentInfo)
	var records [][]string

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		records = append(records, record)

		if len(record) < 11 {
			continue
		}

		factionID := strings.TrimSpace(record[0])
		detachment := strings.TrimSpace(record[8])
		detachmentID := strings.TrimSpace(record[9])

		if detachmentID != "" && detachment != "" {
			detachmentMap[detachmentID] = detachmentInfo{
				ID:        detachmentID,
				FactionID: factionID,
				Name:      detachment,
			}
		}
	}

	// Insert detachments
	detachmentCount := 0
	for _, d := range detachmentMap {
		_, err := pool.Exec(ctx,
			`INSERT INTO detachments (id, faction_id, name) VALUES ($1, $2, $3)
			 ON CONFLICT (id) DO UPDATE SET name = $3`,
			d.ID, d.FactionID, d.Name)
		if err != nil {
			fmt.Printf("Warning: error inserting detachment %s: %v\n", d.ID, err)
			continue
		}
		detachmentCount++
	}

	// Second pass: insert stratagems
	stratagemCount := 0
	for _, record := range records {
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

		_, err := pool.Exec(ctx,
			`INSERT INTO stratagems (id, faction_id, detachment_id, name, type, cp_cost, legend, turn, phase, description)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 ON CONFLICT (id) DO UPDATE SET
			   faction_id = $2, detachment_id = $3, name = $4, type = $5,
			   cp_cost = $6, legend = $7, turn = $8, phase = $9, description = $10`,
			id, factionIDPtr, detachmentIDPtr, name, stratType, cpCost, legend, turn, phase, description)
		if err != nil {
			fmt.Printf("Warning: error inserting stratagem %s: %v\n", id, err)
			continue
		}
		stratagemCount++
	}

	return detachmentCount, stratagemCount, nil
}
