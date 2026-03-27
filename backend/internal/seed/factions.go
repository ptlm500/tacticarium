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

func SeedFactions(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("opening factions file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = '|'
	reader.LazyQuotes = true

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
			return count, fmt.Errorf("reading row: %w", err)
		}

		if len(record) < 3 {
			continue
		}

		id := strings.TrimSpace(record[0])
		name := strings.TrimSpace(record[1])
		link := strings.TrimSpace(record[2])

		if id == "" {
			continue
		}

		_, err = pool.Exec(ctx,
			`INSERT INTO factions (id, name, wahapedia_link) VALUES ($1, $2, $3)
			 ON CONFLICT (id) DO UPDATE SET name = $2, wahapedia_link = $3`,
			id, name, link)
		if err != nil {
			return count, fmt.Errorf("inserting faction %s: %w", id, err)
		}
		count++
	}

	return count, nil
}
