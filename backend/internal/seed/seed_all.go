package seed

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Stats summarises a full reference-data seed run.
type Stats struct {
	Factions           int
	Detachments        int
	Stratagems         int
	ForceDispositions  int
	Missions           int
	MissionMatchups    int
	Cards              int
	DeploymentPatterns int
}

// SeedAll loads the entire vendored 40kdc-data snapshot from dataDir (the
// directory containing the top-level *.json files and a factions/ subdirectory
// with one directory per faction). Entities are seeded in dependency order.
func SeedAll(ctx context.Context, pool *pgxpool.Pool, dataDir string) (Stats, error) {
	var stats Stats

	top := func(name string) string { return filepath.Join(dataDir, name) }

	// Per-faction reference data first (factions -> detachments -> stratagems),
	// since core stratagems and nothing else depend on it being present, and
	// detachments must exist before stratagem faction_id resolution.
	factionDirs, err := filepath.Glob(filepath.Join(dataDir, "factions", "*"))
	if err != nil {
		return stats, fmt.Errorf("globbing faction dirs: %w", err)
	}
	sort.Strings(factionDirs)
	for _, dir := range factionDirs {
		if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
			continue
		}
		if n, err := seedFileIfPresent(ctx, pool, filepath.Join(dir, "factions.json"), SeedFactions); err != nil {
			return stats, err
		} else {
			stats.Factions += n
		}
	}
	// Detachments after ALL factions exist (a detachment could reference any
	// faction, e.g. allied/agents content).
	for _, dir := range factionDirs {
		if n, err := seedFileIfPresent(ctx, pool, filepath.Join(dir, "detachments.json"), SeedDetachments); err != nil {
			return stats, err
		} else {
			stats.Detachments += n
		}
	}
	// Faction stratagems after all detachments exist (faction_id resolves via
	// the detachment lookup).
	for _, dir := range factionDirs {
		if n, err := seedFileIfPresent(ctx, pool, filepath.Join(dir, "stratagems.json"), SeedStratagems); err != nil {
			return stats, err
		} else {
			stats.Stratagems += n
		}
	}

	// Top-level shared data.
	if n, err := seedFileIfPresent(ctx, pool, top("stratagems.json"), SeedStratagems); err != nil {
		return stats, err
	} else {
		stats.Stratagems += n
	}
	if n, err := seedFileIfPresent(ctx, pool, top("force-dispositions.json"), SeedForceDispositions); err != nil {
		return stats, err
	} else {
		stats.ForceDispositions += n
	}
	if n, err := seedFileIfPresent(ctx, pool, top("deployment-patterns.json"), SeedDeploymentPatterns); err != nil {
		return stats, err
	} else {
		stats.DeploymentPatterns += n
	}
	if n, err := seedFileIfPresent(ctx, pool, top("missions.json"), SeedMissions); err != nil {
		return stats, err
	} else {
		stats.Missions += n
	}
	// Matchups depend on dispositions + missions.
	if n, err := seedFileIfPresent(ctx, pool, top("mission-matchups.json"), SeedMissionMatchups); err != nil {
		return stats, err
	} else {
		stats.MissionMatchups += n
	}
	if n, err := seedFileIfPresent(ctx, pool, top("secondary-cards.json"), SeedCards); err != nil {
		return stats, err
	} else {
		stats.Cards += n
	}

	return stats, nil
}

// seedFileIfPresent runs fn against path, treating a missing file as a no-op.
func seedFileIfPresent(
	ctx context.Context,
	pool *pgxpool.Pool,
	path string,
	fn func(context.Context, *pgxpool.Pool, string) (int, error),
) (int, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("stat %s: %w", path, err)
	}
	return fn(ctx, pool, path)
}
