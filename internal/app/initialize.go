package app

import (
	"context"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
	"github.com/th1enq/es-demo/config"
	"github.com/th1enq/es-demo/internal/delivery/http"
	mongo_subscription "github.com/th1enq/es-demo/internal/delivery/kafka/mongo_subcription"
	"github.com/th1enq/es-demo/internal/domain"
	"github.com/th1enq/es-demo/internal/projection"
	"github.com/th1enq/es-demo/internal/repository"
	"github.com/th1enq/es-demo/internal/service"
	"github.com/th1enq/es-demo/internal/utils"
	serviceErrors "github.com/th1enq/es-demo/pkg/errors"
	"github.com/th1enq/es-demo/pkg/es"
	kafkaClient "github.com/th1enq/es-demo/pkg/kafka"
	"github.com/th1enq/es-demo/pkg/logger"
	"github.com/th1enq/es-demo/pkg/mongodb"
	"github.com/th1enq/es-demo/pkg/postgres"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const (
	mongoBankAccountsCollection = "bank_accounts"
)

func Initialize(ctx context.Context) (*Application, error) {
	cfg := config.Load()

	logger, err := logger.Load(cfg.Logger)
	if err != nil {
		return nil, err
	}

	pgx, err := postgres.NewPgxConn(cfg.Postgres)
	if err != nil {
		logger.Error("Failed to connect to Postgres", zap.Error(err))
		return nil, err
	}
	logger.Info("Success connect to Postgres", zap.String("host", cfg.Postgres.Host), zap.Int("port", cfg.Postgres.Port))

	// Run event store migrations
	err = postgres.RunMigrations(ctx, pgx, logger)
	if err != nil {
		logger.Error("Failed to run event store migrations", zap.Error(err))
		return nil, err
	}

	mongodb, err := mongodb.NewMongoDBConn(
		ctx,
		&cfg.MongoDB,
	)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", zap.Error(err))
		return nil, err
	}
	logger.Info("Success connect to MongoDB", zap.String("uri", cfg.MongoDB.URI))

	// init mongodb collection
	err = mongodb.Database(cfg.MongoDB.Db).CreateCollection(ctx, mongoBankAccountsCollection)
	if err != nil {
		if !utils.CheckErrForMessagesCaseInSensitive(err, serviceErrors.ErrMsgMongoCollectionAlreadyExists) {
			logger.Warn("Create Collection Failed, Collection Already Exist", zap.Error(err))
		}
	}

	aggregateIdIndexOptions := options.Index().
		SetSparse(true).
		SetUnique(true)

	aggregateIdIndex, err := mongodb.Database(cfg.MongoDB.Db).
		Collection(mongoBankAccountsCollection).
		Indexes().
		CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "aggregate_id", Value: 1}},
			Options: aggregateIdIndexOptions,
		})

	if err != nil && !utils.CheckErrForMessagesCaseInSensitive(err, serviceErrors.ErrMsgAlreadyExists) {
		logger.Warn("Create One Collection Failed, Collection Already exist", zap.Error(err))
	}
	logger.Info("Created Index on MongoDB", zap.String("index", aggregateIdIndex))

	list, err := mongodb.Database(cfg.MongoDB.Db).Collection(mongoBankAccountsCollection).
		Indexes().
		List(ctx)
	if err != nil {
		logger.Warn("Failed to list indexes", zap.Error(err))
	}

	if list != nil {
		var results []bson.M
		if err := list.All(ctx, &results); err != nil {
			logger.Warn("Failed to list indexes", zap.Error(err))
		}
		logger.Info("Indexes:", zap.Any("results", results))
	}

	collections, err := mongodb.Database(cfg.MongoDB.Db).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		logger.Warn("Failed to list collections", zap.Error(err))
	}
	logger.Info("Created collections:", zap.Any("collections", collections))

	kafkaConn, err := kafkaClient.NewKafkaConn(
		ctx,
		cfg.Kafka,
	)
	if err != nil {
		logger.Error("Failed to connect to Kafka", zap.Error(err))
		return nil, err
	}

	brokers, err := kafkaConn.Brokers()
	if err != nil {
		logger.Error("Failed to get Kafka brokers", zap.Error(err))
		return nil, err
	}
	logger.Info("Success connect to Kafka", zap.Any("brokers", brokers))

	if cfg.Kafka.InitTopics {
		controller, err := kafkaConn.Controller()
		if err != nil {
			logger.Error("Failed to get Kafka controller", zap.Error(err))
			return nil, err
		}

		controllerURI := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
		logger.Info("Kafka controller uri", zap.String("controller URI", controllerURI))

		conn, err := kafka.DialContext(ctx, "tcp", controllerURI)
		if err != nil {
			logger.Error("Failed to dial Kafka controller", zap.Error(err))
			return nil, err
		}
		defer conn.Close()

		bankAccountAggregateTopic := es.GetKafkaAggregateTypeTopic(cfg.KafkaPublisherConfig, string(domain.BankAccountAggregateType))

		if err := conn.CreateTopics(bankAccountAggregateTopic); err != nil {
			logger.Error("Failed to create Kafka topics", zap.Error(err))
			return nil, err
		}

		logger.Info("Kafka topics created successfully")
	}

	serializer := domain.NewEventSerializer()

	kafkaProducer := kafkaClient.NewProducer(
		logger,
		cfg.Kafka.Brokers,
	)
	// Don't close producer here, it will be used throughout the application lifecycle

	eventBus := es.NewKafkaEventsBus(
		kafkaProducer,
		cfg.KafkaPublisherConfig,
	)

	esStore := es.NewPgEventStore(
		cfg.PgStore,
		pgx,
		serializer,
		logger,
		eventBus,
	)

	mongoRepository := repository.NewBankAccountMongoRepository(
		cfg,
		mongodb,
		logger,
	)

	bankService := service.NewBankAccountService(
		logger,
		esStore,
		serializer,
		mongoRepository,
	)

	controller := http.NewController(
		bankService,
	)

	httpServer := http.NewHTTPServer(
		cfg.Server,
		controller,
		logger,
	)

	mongoProjection := projection.NewBankAccountMongoProjection(
		serializer,
		mongoRepository,
		logger,
	)

	mongoSubscription := mongo_subscription.NewBankAccountMongoSubscription(
		logger,
		cfg,
		bankService,
		mongoProjection,
		serializer,
		mongoRepository,
		esStore,
		eventBus,
	)

	mongoConsumerGroup := kafkaClient.NewConsumerGroup(
		cfg.Kafka.Brokers,
		"bank_account_mongo_subscription_group",
		mongoSubscription.ProcessMessagesErrGroup,
		logger,
	)

	return NewApplication(
		cfg,
		httpServer,
		mongoConsumerGroup,
		logger,
	), nil
}
