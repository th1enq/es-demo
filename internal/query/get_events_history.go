package query

import (
	"context"

	"github.com/th1enq/es-demo/internal/dto"
	"github.com/th1enq/es-demo/pkg/es"
	"go.uber.org/zap"
)

type GetEventsHistoryQuery struct {
	AggregateID string `json:"aggregate_id" validate:"required,uuid"`
}

type GetEventsHistoryQueryHandler struct {
	aggregateStore es.AggregateStore
	log            *zap.Logger
}

func NewGetEventsHistoryQueryHandler(
	aggregateStore es.AggregateStore,
	log *zap.Logger,
) *GetEventsHistoryQueryHandler {
	return &GetEventsHistoryQueryHandler{
		aggregateStore: aggregateStore,
		log:            log,
	}
}

func (h *GetEventsHistoryQueryHandler) Handle(
	ctx context.Context,
	query GetEventsHistoryQuery,
) (*dto.EventsHistoryResponse, error) {
	h.log.Info("GetEventsHistoryQueryHandler.Handle", zap.String("aggregateID", query.AggregateID))

	// Load all events from event store
	events, err := h.aggregateStore.LoadEvents(ctx, query.AggregateID)
	if err != nil {
		h.log.Error("Failed to load events", zap.Error(err))
		return nil, err
	}

	// Convert to response format
	eventResponses := make([]dto.EventResponse, 0, len(events))
	for _, event := range events {
		var data interface{}
		var metadata interface{}

		// Parse data if exists
		if len(event.Data) > 0 {
			if err := event.GetJsonData(&data); err != nil {
				h.log.Warn("Failed to parse event data", zap.Error(err), zap.String("eventID", event.EventID))
				data = string(event.Data) // Fallback to raw string
			}
		}

		// Parse metadata if exists
		if len(event.Metadata) > 0 {
			if err := event.GetJsonMetadata(&metadata); err != nil {
				h.log.Warn("Failed to parse event metadata", zap.Error(err), zap.String("eventID", event.EventID))
				metadata = string(event.Metadata) // Fallback to raw string
			}
		}

		eventResponses = append(eventResponses, dto.EventResponse{
			EventID:       event.EventID,
			AggregateID:   event.AggregateID,
			EventType:     string(event.EventType),
			AggregateType: string(event.AggregateType),
			Version:       event.Version,
			Data:          data,
			Metadata:      metadata,
			Timestamp:     event.Timestamp,
		})
	}

	response := &dto.EventsHistoryResponse{
		AggregateID: query.AggregateID,
		TotalEvents: len(eventResponses),
		Events:      eventResponses,
	}

	h.log.Info("GetEventsHistoryQueryHandler.Handle completed",
		zap.String("aggregateID", query.AggregateID),
		zap.Int("totalEvents", len(eventResponses)),
	)

	return response, nil
}
