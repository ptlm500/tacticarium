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

	// The 40kdc snapshot is the sole authority for reference data: clear the
	// tables it owns before repopulating so a reseed can't leave stale rows
	// behind (e.g. 10e detachments carried over the edition reset, which only
	// added columns rather than wiping data). Upsert-only seeding never removes
	// entities that disappear upstream.
	if err := clearReferenceData(ctx, pool); err != nil {
		return stats, err
	}

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

// clearReferenceData empties the 40kdc-owned reference tables in FK-safe order
// (children before parents). game_players carries faction_id/detachment_id FKs
// but the 11e code leaves them NULL — faction/detachment live in the game-state
// JSONB — so deleting factions/detachments does not violate them.
func clearReferenceData(ctx context.Context, pool *pgxpool.Pool) error {
	tables := []string{
		"mission_matchups", // -> force_dispositions, primary_missions
		"stratagems_11e",   // -> factions, detachments
		"detachments",      // -> factions
		"factions",
		"force_dispositions",
		"primary_missions",
		"cards",
		"deployment_patterns",
	}
	for _, t := range tables {
		if _, err := pool.Exec(ctx, "DELETE FROM "+t); err != nil {
			return fmt.Errorf("clearing %s: %w", t, err)
		}
	}
	return nil
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
