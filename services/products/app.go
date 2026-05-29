package products

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"uni-test-repo/pkg/logger"
	"uni-test-repo/pkg/migrations"
	"uni-test-repo/pkg/postgres"
	"uni-test-repo/services/products/config"
	"uni-test-repo/services/products/internal/outbox"
	"uni-test-repo/services/products/internal/product"
	"uni-test-repo/services/products/internal/product/productcontroller"
	"uni-test-repo/services/products/internal/product/productrepo"
)

//go:embed migrations/*.sql
var MIGRATION_FS embed.FS

func Run(cfg config.Config) {
	logger.Setup(logger.Options{
		Level:   cfg.LogLevel,
		Console: strings.ToLower(os.Getenv("LOG_FORMAT")) == "console",
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pg, err := postgres.New(ctx, cfg.PgURL, cfg.PgPoolMax)
	if err != nil {
		slog.Error("Failed to connect to postgres", slog.Any("error", err))
		os.Exit(1)
	}
	defer pg.Close()

	if err := migrations.ApplyMigrations(cfg.PgURL, MIGRATION_FS); err != nil {
		slog.Error("Failed to apply migrations", slog.Any("error", err))
		os.Exit(1)
	}

	txOutbox := outbox.TxStoreFactory(pg.Builder)

	productRepo := productrepo.New(pg)
	productService := product.NewProductService(
		productRepo,
		pg,
		productrepo.TxRepoFactory(pg.Builder),
		txOutbox,
	)
	productH := productcontroller.NewHTTPHandler(productService)

	outboxWorker := outbox.NewWorker(pg, txOutbox, outbox.LogPublisher{}, outbox.WorkerConfig{
		PollInterval: cfg.OutboxPollInterval,
		BatchSize:    cfg.OutboxBatchSize,
	})
	go outboxWorker.Start(ctx)

	engine := NewGinEngine()
	router := NewRouter(productH)
	router.SetUp(engine)

	go func() {
		slog.Info("Starting Products HTTP server", "port", cfg.Port)
		if err := engine.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
			slog.Error("HTTP server error", slog.Any("error", err))
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down Products service gracefully...")
}
