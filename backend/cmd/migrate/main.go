package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/dzulfikarq/kostify/backend/internal/config"
	infrapostgres "github.com/dzulfikarq/kostify/backend/internal/infrastructure/postgres"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	downAll := flag.Bool("down-all", false, "rollback seluruh migration")
	steps := flag.Int("steps", 1, "jumlah langkah down")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fatal("config tidak valid: %v", err)
	}

	direction := "up"
	if *downAll {
		direction = "down"
	} else if flag.NArg() > 0 {
		direction = flag.Arg(0)
	}

	switch direction {
	case "up":
		if err := infrapostgres.Migrate(cfg.MigrateURL(), cfg.MigrationsDir, true, -1); err != nil {
			fatal("migrate up gagal: %v", err)
		}
		fmt.Println("migration up selesai")
	case "down":
		n := *steps
		if flag.NArg() > 1 {
			parsed, err := strconv.Atoi(flag.Arg(1))
			if err != nil {
				fatal("jumlah step tidak valid: %v", err)
			}
			n = parsed
		}
		if n < 0 {
			if err := infrapostgres.Migrate(cfg.MigrateURL(), cfg.MigrationsDir, false, -1); err != nil {
				fatal("migrate down-all gagal: %v", err)
			}
			fmt.Println("migration down-all selesai")
			return
		}
		if err := infrapostgres.Migrate(cfg.MigrateURL(), cfg.MigrationsDir, false, n); err != nil {
			fatal("migrate down gagal: %v", err)
		}
		fmt.Printf("migration down %d selesai\n", n)
	default:
		fatal("direction tidak dikenal: %s (gunakan up|down)", direction)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
