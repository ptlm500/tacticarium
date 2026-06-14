package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/peter/tacticarium/backend/internal/db"
	"github.com/peter/tacticarium/backend/internal/seed"
)

func main() {
	dataDir := flag.String("data", "", "Path to the vendored 40kdc-data directory")
	all := flag.Bool("all", false, "Seed all reference data (uses the default data dir)")
	migrate := flag.Bool("migrate", false, "Run database migrations before seeding")
	flag.Parse()

	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/tacticarium?sslmode=disable"
	}

	ctx := context.Background()

	pool, err := db.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if *migrate {
		log.Println("Running migrations...")
		if err := db.RunMigrations(dbURL); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations complete.")
	}

	if *all && *dataDir == "" {
		// Default path relative to the backend working directory.
		*dataDir = "../data/40kdc"
	}

	if *dataDir != "" {
		log.Printf("Seeding 40kdc-data from %s...", *dataDir)
		stats, err := seed.SeedAll(ctx, pool, *dataDir)
		if err != nil {
			log.Fatalf("Failed to seed reference data: %v", err)
		}
		fmt.Printf(
			"Seeded: %d factions, %d detachments, %d stratagems, %d force dispositions, %d missions, %d matchups, %d cards, %d deployment patterns\n",
			stats.Factions, stats.Detachments, stats.Stratagems, stats.ForceDispositions,
			stats.Missions, stats.MissionMatchups, stats.Cards, stats.DeploymentPatterns,
		)
	}

	if *dataDir == "" && !*migrate {
		fmt.Println("Usage: seed [--migrate] [--data path | --all]")
		os.Exit(1)
	}

	log.Println("Seeding complete!")
}
