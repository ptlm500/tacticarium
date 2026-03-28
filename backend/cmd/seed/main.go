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
	factionsFile := flag.String("factions", "", "Path to Factions.csv")
	stratagemFile := flag.String("stratagems", "", "Path to Stratagems.csv")
	missionsFile := flag.String("missions", "", "Path to missions.json")
	all := flag.Bool("all", false, "Seed all data (uses default paths)")
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
		if err := db.RunMigrations(ctx, pool); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations complete.")
	}

	if *all {
		if *factionsFile == "" {
			*factionsFile = "../Factions.csv"
		}
		if *stratagemFile == "" {
			*stratagemFile = "../Stratagems.csv"
		}
		if *missionsFile == "" {
			*missionsFile = "../missions.json"
		}
	}

	if *factionsFile != "" {
		log.Printf("Seeding factions from %s...", *factionsFile)
		count, err := seed.SeedFactions(ctx, pool, *factionsFile)
		if err != nil {
			log.Fatalf("Failed to seed factions: %v", err)
		}
		fmt.Printf("Seeded %d factions\n", count)
	}

	if *stratagemFile != "" {
		log.Printf("Seeding stratagems from %s...", *stratagemFile)
		detCount, stratCount, err := seed.SeedStratagems(ctx, pool, *stratagemFile)
		if err != nil {
			log.Fatalf("Failed to seed stratagems: %v", err)
		}
		fmt.Printf("Seeded %d detachments, %d stratagems\n", detCount, stratCount)
	}

	if *missionsFile != "" {
		log.Printf("Seeding missions from %s...", *missionsFile)
		stats, err := seed.SeedMissions(ctx, pool, *missionsFile)
		if err != nil {
			log.Fatalf("Failed to seed missions: %v", err)
		}
		fmt.Printf("Seeded %d missions, %d mission rules, %d secondaries, %d challenger cards, %d gambits\n",
			stats.Missions, stats.MissionRules, stats.Secondaries, stats.ChallengerCards, stats.Gambits)
	}

	if *factionsFile == "" && *stratagemFile == "" && *missionsFile == "" && !*migrate {
		fmt.Println("Usage: seed [--migrate] [--factions path] [--stratagems path] [--missions path] [--all]")
		os.Exit(1)
	}

	log.Println("Seeding complete!")
}
