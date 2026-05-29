package notifications

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"uni-test-repo/pkg/logger"
	"uni-test-repo/services/notifications/config"
	"uni-test-repo/services/notifications/internal/consumer"
	"uni-test-repo/services/notifications/internal/notifier"
)

func Run(cfg config.Config) {
	logger.Setup(logger.Options{
		Level:   cfg.LogLevel,
		Console: strings.ToLower(os.Getenv("LOG_FORMAT")) == "console",
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	c := consumer.New(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.GroupID)
	defer func() {
		if err := c.Close(); err != nil {
			slog.Error("Failed to close consumer", slog.Any("error", err))
		}
	}()

	n := notifier.New()

	slog.Info("Notifications service started", "topic", cfg.KafkaTopic, "group", cfg.GroupID)
	if err := c.Start(ctx, n.Handle); err != nil {
		slog.Error("Consumer stopped with error", slog.Any("error", err))
	}
	slog.Info("Shutting down Notifications service gracefully...")
}
