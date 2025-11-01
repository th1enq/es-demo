package domain

import "context"

type UpdateProjectionCallback func(projection *BankAccountMongoProjection) *BankAccountMongoProjection

type MongoRepository interface {
	Insert(ctx context.Context, projection *BankAccountMongoProjection) error
	Update(ctx context.Context, projection *BankAccountMongoProjection) error
	Upsert(ctx context.Context, projection *BankAccountMongoProjection) error

	DeleteByAggregateID(ctx context.Context, aggregateID string) error
	UpdateConcurrently(ctx context.Context, aggregateID string, updateCb UpdateProjectionCallback, expectedVersion uint64) error
	GetByAggregateID(ctx context.Context, aggregateID string) (*BankAccountMongoProjection, error)
	GetByEmail(ctx context.Context, email string) (*BankAccountMongoProjection, error)
}
