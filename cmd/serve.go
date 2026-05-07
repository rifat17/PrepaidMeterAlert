package cmd

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/m4hi2/MeterAlertBot/internal/config"
	"github.com/m4hi2/MeterAlertBot/internal/database/repo"
	"github.com/m4hi2/MeterAlertBot/internal/datasources/desco"
	"github.com/muesli/coral"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
)

var serveCmd = &coral.Command{
	Use:   "serve",
	Short: "Run the alert daemon",
	RunE:  runServe,
}

func runServe(cmd *coral.Command, _ []string) error {
	db, err := openDB(config.Get().DB)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = repo.NewUserRepo(db)
	_ = repo.NewMeterRepo(db)
	_ = repo.NewProviderRepo(db)
	_ = repo.NewNotificationLogRepo(db)

	_ = desco.NewService(config.Get().Desco)

	slog.InfoContext(ctx, "meterbot started")

	<-ctx.Done()
	slog.Info("shutting down")
	return nil
}

func openDB(cfg config.DBConfig) (*bun.DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.DSN)))
	db := bun.NewDB(sqldb, nil)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	slog.Info("database connected")
	return db, nil
}
