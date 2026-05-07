package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/m4hi2/MeterAlertBot/internal/config"
	"github.com/m4hi2/MeterAlertBot/internal/database"
	"github.com/m4hi2/MeterAlertBot/internal/database/repo"
	"github.com/m4hi2/MeterAlertBot/internal/datasources/desco"
	"github.com/muesli/coral"
)

var serveCmd = &coral.Command{
	Use:   "serve",
	Short: "Run the alert daemon",
	RunE:  runServe,
}

func runServe(cmd *coral.Command, _ []string) error {
	db, err := database.Open(config.Get().DB)
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
