package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/th1enq/es-demo/internal/domain"
	"github.com/th1enq/es-demo/internal/events"
	"github.com/th1enq/es-demo/pkg/es"
	"go.uber.org/zap"
)

const (
	BankAccountIndexName = "bank_accounts"
)

// ReplayService handles replaying events from PostgreSQL to Elasticsearch
type ReplayService struct {
	eventStore es.EventStore
	esRepo     domain.ElasticsearchRepository
	serializer es.Serializer
	logger     *zap.Logger
}

// NewReplayService creates a new replay service
func NewReplayService(
	eventStore es.EventStore,
	esRepo domain.ElasticsearchRepository,
	serializer es.Serializer,
	logger *zap.Logger,
) *ReplayService {
	return &ReplayService{
		eventStore: eventStore,
		esRepo:     esRepo,
		serializer: serializer,
		logger:     logger,
	}
}

// ReplayResult contains the result of a replay operation
type ReplayResult struct {
	TotalEvents      int              `json:"totalEvents"`
	ProcessedEvents  int              `json:"processedEvents"`
	CreatedAccounts  int              `json:"createdAccounts"`
	UpdatedAccounts  int              `json:"updatedAccounts"`
	Errors           []string         `json:"errors,omitempty"`
	Duration         time.Duration    `json:"duration"`
	AccountSummaries []AccountSummary `json:"accountSummaries"`
}

// AccountSummary provides a summary of each account after replay
type AccountSummary struct {
	AggregateID      string    `json:"aggregateId"`
	Email            string    `json:"email"`
	FullName         string    `json:"fullName"`
	Balance          int64     `json:"balance"`
	TransactionCount int       `json:"transactionCount"`
	LastActivity     time.Time `json:"lastActivity"`
}

// ReplayAllEvents replays all events from PostgreSQL to Elasticsearch
func (s *ReplayService) ReplayAllEvents(ctx context.Context, recreateIndex bool) (*ReplayResult, error) {
	startTime := time.Now()
	result := &ReplayResult{
		Errors:           make([]string, 0),
		AccountSummaries: make([]AccountSummary, 0),
	}

	s.logger.Info("Starting event replay to Elasticsearch", zap.Bool("recreate_index", recreateIndex))

	// Recreate index if requested
	if recreateIndex {
		if err := s.esRepo.DeleteIndex(BankAccountIndexName); err != nil {
			s.logger.Warn("Failed to delete existing index (may not exist)", zap.Error(err))
		}

		if err := s.esRepo.CreateIndex(BankAccountIndexName); err != nil {
			s.logger.Error("Failed to create index", zap.Error(err))
			return nil, errors.Wrap(err, "failed to create index")
		}
		s.logger.Info("Created new Elasticsearch index", zap.String("index", BankAccountIndexName))
	}

	// Get all events from event store
	allEvents, err := s.eventStore.GetAllEvents(ctx)
	if err != nil {
		s.logger.Error("Failed to get all events", zap.Error(err))
		return nil, errors.Wrap(err, "failed to get all events")
	}

	result.TotalEvents = len(allEvents)
	s.logger.Info("Retrieved all events for replay", zap.Int("count", result.TotalEvents))

	// Group events by aggregate ID to maintain projection state
	aggregateProjections := make(map[string]*domain.BankAccountElasticsearchProjection)

	// Process events in order
	for _, event := range allEvents {
		if err := s.processEvent(ctx, event, aggregateProjections); err != nil {
			errMsg := fmt.Sprintf("Failed to process event %s for aggregate %s: %v", event.EventID, event.AggregateID, err)
			result.Errors = append(result.Errors, errMsg)
			s.logger.Error("Failed to process event",
				zap.String("event_id", event.EventID),
				zap.String("aggregate_id", event.AggregateID),
				zap.Error(err))
			continue
		}
		result.ProcessedEvents++
	}

	// Index all projections to Elasticsearch
	documents := make(map[string]interface{})
	for aggregateID, projection := range aggregateProjections {
		documents[aggregateID] = projection

		// Create account summary
		summary := AccountSummary{
			AggregateID:      projection.AggregateID,
			Email:            projection.Email,
			FullName:         projection.GetFullName(),
			Balance:          projection.Balance.Amount,
			TransactionCount: projection.TransactionCount,
			LastActivity:     projection.LastActivity,
		}
		result.AccountSummaries = append(result.AccountSummaries, summary)
	}

	if len(documents) > 0 {
		if err := s.esRepo.BulkIndex(BankAccountIndexName, documents); err != nil {
			s.logger.Error("Failed to bulk index documents", zap.Error(err))
			return nil, errors.Wrap(err, "failed to bulk index documents")
		}
		s.logger.Info("Successfully indexed all projections", zap.Int("count", len(documents)))
	}

	result.CreatedAccounts = len(aggregateProjections)
	result.Duration = time.Since(startTime)

	s.logger.Info("Event replay completed successfully",
		zap.Int("total_events", result.TotalEvents),
		zap.Int("processed_events", result.ProcessedEvents),
		zap.Int("created_accounts", result.CreatedAccounts),
		zap.Duration("duration", result.Duration))

	return result, nil
}

// processEvent processes a single event and updates the appropriate projection
func (s *ReplayService) processEvent(ctx context.Context, event es.Event, projections map[string]*domain.BankAccountElasticsearchProjection) error {
	// Get or create projection for this aggregate
	projection, exists := projections[event.AggregateID]
	if !exists {
		projection = domain.NewBankAccountElasticsearchProjection(event.AggregateID)
		projections[event.AggregateID] = projection
	}

	// Deserialize the event
	deserializedEvent, err := s.serializer.DeserializeEvent(event)
	if err != nil {
		return errors.Wrap(err, "failed to deserialize event")
	}

	// Apply event to projection based on event type
	switch event.EventType {
	case events.BankAccountCreatedEventTypeV1:
		if createEvent, ok := deserializedEvent.(*events.BankAccountCreatedEventV1); ok {
			projection.WhenBankAccountCreated(*createEvent, event.AggregateID, event.Version, event.Timestamp)
		} else {
			return errors.New("failed to cast to BankAccountCreatedEventV1")
		}

	case events.BalancedDepositedEventTypeV1:
		if depositEvent, ok := deserializedEvent.(*events.BalanceDepositedEventV1); ok {
			projection.WhenBalanceDeposited(*depositEvent, event.Version, event.Timestamp)
		} else {
			return errors.New("failed to cast to BalanceDepositedEventV1")
		}

	case events.BalanceWithdrawedEventTypeV1:
		if withdrawEvent, ok := deserializedEvent.(*events.BalanceWithdrawedEventV1); ok {
			projection.WhenBalanceWithdrawn(*withdrawEvent, event.Version, event.Timestamp)
		} else {
			return errors.New("failed to cast to BalanceWithdrawedEventV1")
		}

	default:
		s.logger.Warn("Unknown event type encountered during replay",
			zap.String("event_type", string(event.EventType)),
			zap.String("aggregate_id", event.AggregateID))
		return nil // Skip unknown events, don't fail the entire replay
	}

	return nil
}

// GetAccountByID retrieves a bank account from Elasticsearch
func (s *ReplayService) GetAccountByID(ctx context.Context, aggregateID string) (*domain.BankAccountElasticsearchProjection, error) {
	projection, err := s.esRepo.GetDocument(BankAccountIndexName, aggregateID)
	if err != nil {
		s.logger.Error("Failed to get account from Elasticsearch",
			zap.String("aggregate_id", aggregateID),
			zap.Error(err))
		return nil, errors.Wrap(err, "failed to get account from Elasticsearch")
	}

	return projection, nil
}

// SearchAccounts searches for bank accounts in Elasticsearch
func (s *ReplayService) SearchAccounts(ctx context.Context, query map[string]interface{}) ([]*domain.BankAccountElasticsearchProjection, error) {
	projections, err := s.esRepo.Search(BankAccountIndexName, query)
	if err != nil {
		s.logger.Error("Failed to search accounts in Elasticsearch", zap.Error(err))
		return nil, errors.Wrap(err, "failed to search accounts in Elasticsearch")
	}

	return projections, nil
}

// GetAccountSummary provides analytics summary for accounts
func (s *ReplayService) GetAccountSummary(ctx context.Context) (map[string]interface{}, error) {
	// Get all accounts
	allAccountsQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"size": 1000, // Adjust based on your needs
	}

	accounts, err := s.SearchAccounts(ctx, allAccountsQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all accounts")
	}

	totalBalance := int64(0)
	totalDeposits := int64(0)
	totalWithdrawals := int64(0)
	totalTransactions := 0
	activeAccounts := 0

	for _, account := range accounts {
		totalBalance += account.Balance.Amount
		totalDeposits += account.TotalDeposits
		totalWithdrawals += account.TotalWithdrawals
		totalTransactions += account.TransactionCount
		if account.IsActive() {
			activeAccounts++
		}
	}

	summary := map[string]interface{}{
		"totalAccounts":     len(accounts),
		"activeAccounts":    activeAccounts,
		"totalBalance":      totalBalance,
		"totalDeposits":     totalDeposits,
		"totalWithdrawals":  totalWithdrawals,
		"totalTransactions": totalTransactions,
		"netFlow":           totalDeposits - totalWithdrawals,
		"averageBalance":    float64(totalBalance) / float64(len(accounts)),
	}

	return summary, nil
}

// DeleteElasticsearchIndex deletes the bank accounts index from Elasticsearch
func (s *ReplayService) DeleteElasticsearchIndex(ctx context.Context) error {
	s.logger.Info("Deleting Elasticsearch index", zap.String("index", BankAccountIndexName))

	err := s.esRepo.DeleteIndex(BankAccountIndexName)
	if err != nil {
		s.logger.Error("Failed to delete Elasticsearch index",
			zap.String("index", BankAccountIndexName),
			zap.Error(err))
		return errors.Wrap(err, "failed to delete Elasticsearch index")
	}

	s.logger.Info("Successfully deleted Elasticsearch index", zap.String("index", BankAccountIndexName))
	return nil
}
