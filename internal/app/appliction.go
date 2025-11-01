package app

import (
	"context"
	"syscall"

	"github.com/th1enq/es-demo/config"
	"github.com/th1enq/es-demo/internal/delivery/http"
	"github.com/th1enq/es-demo/internal/domain"
	"github.com/th1enq/es-demo/internal/utils"
	"github.com/th1enq/es-demo/pkg/es"
	kafka_client "github.com/th1enq/es-demo/pkg/kafka"
	"go.uber.org/zap"
)

type Application struct {
	cfg               *config.Config
	server            http.HTTPServer
	mongoSubscription kafka_client.ConsumerGroup
	logger            *zap.Logger
}

func NewApplication(
	cfg *config.Config,
	server http.HTTPServer,
	mongoSubscription kafka_client.ConsumerGroup,
	logger *zap.Logger,
) *Application {
	return &Application{
		cfg:               cfg,
		server:            server,
		mongoSubscription: mongoSubscription,
		logger:            logger,
	}
}

func (app *Application) Start(ctx context.Context) error {
	app.logger.Info("Starting application ...")
	go func() {
		if err := app.server.Start(ctx); err != nil {
			app.logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	topics := []string{
		es.GetTopicName(app.cfg.KafkaPublisherConfig.TopicPrefix, string(domain.BankAccountAggregateType)),
	}
	go func() {
		if err := app.mongoSubscription.ConsumeTopicWithErrGroup(
			ctx,
			topics,
			app.cfg.Projections.MongoSubscriptionPoolSize,
		); err != nil {
			app.logger.Fatal("Failed to start MongoDB subscription consumer group", zap.Error(err))
		}
	}()
	utils.BlockUntilSignal(syscall.SIGINT, syscall.SIGTERM)
	return nil
}
