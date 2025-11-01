package repository

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/th1enq/es-demo/config"
	"github.com/th1enq/es-demo/internal/domain"
	"github.com/th1enq/es-demo/pkg/constants"
	"github.com/th1enq/es-demo/pkg/es"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type bankAccountMongoRepository struct {
	cfg    *config.Config
	db     *mongo.Client
	logger *zap.Logger
}

func NewBankAccountMongoRepository(
	cfg *config.Config,
	db *mongo.Client,
	logger *zap.Logger,
) domain.MongoRepository {
	return &bankAccountMongoRepository{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}
}

// DeleteByAggregateID implements domain.MongoRepository.
func (b *bankAccountMongoRepository) DeleteByAggregateID(ctx context.Context, aggregateID string) error {
	b.logger.Info("Deleting bank account", zap.String("aggregateID", aggregateID))
	filter := bson.M{constants.MongoAggregateID: aggregateID}
	ops := options.Delete()

	_, err := b.bankAccountsCollection().DeleteOne(ctx, filter, ops)
	if err != nil {
		b.logger.Error("MongoDB delete failed", zap.String("aggregateID", aggregateID), zap.Error(err))
		return errors.Wrapf(err, "DeleteByAggregateID [FindOneAndDelete] aggregateID: %s", aggregateID)
	}
	b.logger.Info("Deleted bank account", zap.String("aggregateID", aggregateID))
	return nil
}

// GetByAggregateID implements domain.MongoRepository.
func (b *bankAccountMongoRepository) GetByAggregateID(ctx context.Context, aggregateID string) (*domain.BankAccountMongoProjection, error) {
	b.logger.Info("Getting bank account", zap.String("aggregateID", aggregateID))
	filter := bson.M{constants.MongoAggregateID: aggregateID}
	var projection domain.BankAccountMongoProjection

	err := b.bankAccountsCollection().FindOne(ctx, filter).Decode(&projection)
	if err != nil {
		b.logger.Error("MongoDB find failed", zap.String("aggregateID", aggregateID), zap.Error(err))
		return nil, errors.Wrapf(err, "[FindOne] aggregateID: %s", projection.AggregateID)
	}
	b.logger.Info("Got bank account", zap.String("aggregateID", aggregateID))
	return &projection, nil
}

// GetByEmail implements domain.MongoRepository.
func (b *bankAccountMongoRepository) GetByEmail(ctx context.Context, email string) (*domain.BankAccountMongoProjection, error) {
	b.logger.Info("Getting bank account by email", zap.String("email", email))
	filter := bson.M{"email": email}
	var projection domain.BankAccountMongoProjection

	err := b.bankAccountsCollection().FindOne(ctx, filter).Decode(&projection)
	if err != nil {
		b.logger.Error("MongoDB find by email failed", zap.String("email", email), zap.Error(err))
		return nil, errors.Wrapf(err, "[FindOne] email: %s", email)
	}
	b.logger.Info("Got bank account by email", zap.String("email", email))
	return &projection, nil
}

// Update implements domain.MongoRepository.
func (b *bankAccountMongoRepository) Update(ctx context.Context, projection *domain.BankAccountMongoProjection) error {
	b.logger.Info("Updating bank account", zap.String("aggregateID", projection.AggregateID))
	projection.ID = ""
	projection.UpdatedAt = time.Now().UTC()

	ops := options.FindOneAndUpdate()
	ops.SetReturnDocument(options.After)
	ops.SetUpsert(false)
	filter := bson.M{constants.MongoAggregateID: projection.AggregateID}

	err := b.bankAccountsCollection().FindOneAndUpdate(ctx, filter, bson.M{"$set": projection}, ops).Decode(projection)
	if err != nil {
		b.logger.Error("MongoDB update failed", zap.String("aggregateID", projection.AggregateID), zap.Error(err))
		return errors.Wrapf(err, "[FindOneAndUpdate] aggregateID: %s", projection.AggregateID)
	}
	b.logger.Info("Updated bank account", zap.String("aggregateID", projection.AggregateID))
	return nil
}

// UpdateConcurrently implements domain.MongoRepository.
func (b *bankAccountMongoRepository) UpdateConcurrently(ctx context.Context, aggregateID string, updateCb domain.UpdateProjectionCallback, expectedVersion uint64) error {
	b.logger.Info("Updating bank account concurrently", zap.String("aggregateID", aggregateID), zap.Uint64("expectedVersion", expectedVersion))
	session, err := b.db.StartSession()
	if err != nil {
		b.logger.Error("Failed to start session", zap.String("aggregateID", aggregateID), zap.Uint64("expectedVersion", expectedVersion), zap.Error(err))
		return errors.Wrapf(err, "StartSession aggregateID: %s, expectedVersion: %d", aggregateID, expectedVersion)
	}
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sessionContext mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			b.logger.Error("Failed to start transaction", zap.String("aggregateID", aggregateID), zap.Uint64("expectedVersion", expectedVersion), zap.Error(err))
			return errors.Wrapf(err, "StartTransaction aggregateID: %s, expectedVersion: %d", aggregateID, expectedVersion)
		}

		filter := bson.M{constants.MongoAggregateID: aggregateID}
		foundProjection := &domain.BankAccountMongoProjection{}

		err := b.bankAccountsCollection().FindOne(ctx, filter).Decode(foundProjection)
		if err != nil {
			b.logger.Error("MongoDB find failed", zap.String("aggregateID", aggregateID), zap.Error(err))
			return errors.Wrapf(err, "[FindOne] aggregateID: %s, expectedVersion: %d", aggregateID, expectedVersion)
		}

		if foundProjection.Version != expectedVersion {
			b.logger.Error("Version mismatch", zap.String("aggregateID", aggregateID), zap.Uint64("expectedVersion", expectedVersion), zap.Uint64("actualVersion", foundProjection.Version))
			return errors.Wrapf(es.ErrInvalidEventVersion, "[FindOne] aggregateID: %s, expectedVersion: %d", aggregateID, expectedVersion)
		}

		foundProjection = updateCb(foundProjection)

		foundProjection.ID = ""
		foundProjection.UpdatedAt = time.Now().UTC()

		ops := options.FindOneAndUpdate()
		ops.SetReturnDocument(options.After)
		ops.SetUpsert(false)
		filter = bson.M{constants.MongoAggregateID: foundProjection.AggregateID}

		err = b.bankAccountsCollection().FindOneAndUpdate(ctx, filter, bson.M{"$set": foundProjection}, ops).Decode(foundProjection)
		if err != nil {
			b.logger.Error("MongoDB concurrent update failed", zap.String("aggregateID", foundProjection.AggregateID), zap.Uint64("expectedVersion", expectedVersion), zap.Error(err))
			return errors.Wrapf(err, "[FindOneAndUpdate] aggregateID: %s, expectedVersion: %d", foundProjection.AggregateID, expectedVersion)
		}

		return session.CommitTransaction(ctx)
	})
	if err != nil {
		if err := session.AbortTransaction(ctx); err != nil {
			b.logger.Error("Failed to abort transaction", zap.String("aggregateID", aggregateID), zap.Uint64("expectedVersion", expectedVersion), zap.Error(err))
			return errors.Wrapf(err, "AbortTransaction aggregateID: %s, expectedVersion: %d", aggregateID, expectedVersion)
		}
		b.logger.Error("Failed to update concurrently", zap.String("aggregateID", aggregateID), zap.Uint64("expectedVersion", expectedVersion), zap.Error(err))
		return errors.Wrapf(err, "mongo.WithSession aggregateID: %s, expectedVersion: %d", aggregateID, expectedVersion)
	}
	b.logger.Info("Updated bank account concurrently", zap.String("aggregateID", aggregateID), zap.Uint64("expectedVersion", expectedVersion))
	return nil
}

// Upsert implements domain.MongoRepository.
func (b *bankAccountMongoRepository) Upsert(ctx context.Context, projection *domain.BankAccountMongoProjection) error {
	projection.UpdatedAt = time.Now().UTC()

	ops := options.FindOneAndUpdate()
	ops.SetReturnDocument(options.After)
	ops.SetUpsert(true)
	filter := bson.M{constants.MongoAggregateID: projection.AggregateID}

	err := b.bankAccountsCollection().FindOneAndUpdate(ctx, filter, bson.M{"$set": projection}, ops).Decode(projection)
	if err != nil {
		return errors.Wrapf(err, "Upsert [FindOneAndUpdate] aggregateID: %s", projection.AggregateID)
	}

	return nil
}

func (b *bankAccountMongoRepository) Insert(ctx context.Context, projection *domain.BankAccountMongoProjection) error {
	b.logger.Info("Inserting bank account", zap.String("aggregateID", projection.AggregateID))
	_, err := b.bankAccountsCollection().InsertOne(ctx, projection)
	if err != nil {
		b.logger.Error("MongoDB insert failed", zap.String("aggregateID", projection.AggregateID), zap.Error(err))
		return errors.Wrapf(err, "[InsertOne] AggregateID: %s", projection.AggregateID)
	}
	b.logger.Info("Inserted bank account", zap.String("aggregateID", projection.AggregateID))
	return nil
}

func (b *bankAccountMongoRepository) bankAccountsCollection() *mongo.Collection {
	return b.db.Database(b.cfg.MongoDB.Db).Collection("bank_accounts")
}
