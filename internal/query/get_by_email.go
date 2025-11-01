package query

import (
	"context"

	"github.com/pkg/errors"
	"github.com/th1enq/es-demo/internal/domain"
	bankAccountErrors "github.com/th1enq/es-demo/internal/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type GetBankAccountByEmailQuery struct {
	Email string `json:"email" validate:"required,email"`
}

type GetBankAccountByEmail interface {
	Handle(ctx context.Context, query GetBankAccountByEmailQuery) (*domain.BankAccountMongoProjection, error)
}

type getBankAccountByEmailQuery struct {
	mongoRepository domain.MongoRepository
	logger          *zap.Logger
}

func NewGetBankAccountByEmailQuery(
	bankAccountRepo domain.MongoRepository,
	logger *zap.Logger,
) GetBankAccountByEmail {
	return &getBankAccountByEmailQuery{
		mongoRepository: bankAccountRepo,
		logger:          logger,
	}
}

func (q *getBankAccountByEmailQuery) Handle(ctx context.Context, query GetBankAccountByEmailQuery) (*domain.BankAccountMongoProjection, error) {
	q.logger.Info("GetBankAccountByEmail query", zap.String("email", query.Email))

	projection, err := q.mongoRepository.GetByEmail(ctx, query.Email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.Wrapf(bankAccountErrors.ErrBankAccountNotFound, "email: %s", query.Email)
		}
		return nil, err
	}

	return projection, nil
}
