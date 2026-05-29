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

func bootstrap(cfg config.Config) (context.Context, context.CancelFunc, *postgres.Postgres) {
	logger.Setup(logger.Options{
		Level:   cfg.LogLevel,
		Console: strings.ToLower(os.Getenv("LOG_FORMAT")) == "console",
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	pg, err := postgres.New(ctx, cfg.PgURL, cfg.PgPoolMax)
	if err != nil {
		slog.Error("Failed to connect to postgres", slog.Any("error", err))
		os.Exit(1)
	}
	return ctx, cancel, pg
}

func RunAPI(cfg config.Config) {
	ctx, cancel, pg := bootstrap(cfg)
	defer cancel()
	defer pg.Close()

	if err := migrations.ApplyMigrations(cfg.PgURL, MIGRATION_FS); err != nil {
		slog.Error("Failed to apply migrations", slog.Any("error", err))
		os.Exit(1)
	}

	productRepo := productrepo.New(pg)
	productService := product.NewProductService(
		productRepo,
		pg,
		productrepo.TxRepoFactory(pg.Builder),
		outbox.TxStoreFactory(pg.Builder),
	)
	productH := productcontroller.NewHTTPHandler(productService)

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
	slog.Info("Shutting down Products API gracefully...")
}

func RunWorker(cfg config.Config) {
	ctx, cancel, pg := bootstrap(cfg)
	defer cancel()
	defer pg.Close()

	publisher := outbox.NewKafkaPublisher(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer func() {
		if err := publisher.Close(); err != nil {
			slog.Error("Failed to close Kafka publisher", slog.Any("error", err))
		}
	}()

	worker := outbox.NewWorker(pg, outbox.TxStoreFactory(pg.Builder), publisher, outbox.WorkerConfig{
		PollInterval: cfg.OutboxPollInterval,
		BatchSize:    cfg.OutboxBatchSize,
	})

	worker.Start(ctx) // blocks until ctx is cancelled
	slog.Info("Shutting down Products worker gracefully...")
}
