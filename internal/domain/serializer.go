package domain

import (
	"github.com/pkg/errors"
	"github.com/th1enq/es-demo/internal/events"
	"github.com/th1enq/es-demo/pkg/es"
	"github.com/th1enq/es-demo/pkg/es/serializer"
)

var (
	ErrInvalidEvent = errors.New("invalid event")
)

type eventSerializer struct {
}

func NewEventSerializer() *eventSerializer {
	return &eventSerializer{}
}

func (s *eventSerializer) SerializeEvent(aggregate es.Aggregate, event any) (es.Event, error) {
	eventsBytes, err := serializer.Marshal(event)
	if err != nil {
		return es.Event{}, errors.Wrapf(err, "serializer.Marshal aggregateID: %s", aggregate.GetID())
	}

	switch evt := event.(type) {
	case *events.BankAccountCreatedEventV1:
		return es.NewEvent(aggregate, events.BankAccountCreatedEventTypeV1, eventsBytes, evt.Metadata), nil
	case *events.BalanceDepositedEventV1:
		return es.NewEvent(aggregate, events.BalancedDepositedEventTypeV1, eventsBytes, evt.Metadata), nil
	case *events.BalanceWithdrawedEventV1:
		return es.NewEvent(aggregate, events.BalanceWithdrawedEventTypeV1, eventsBytes, evt.Metadata), nil
	default:
		return es.Event{}, errors.Wrapf(ErrInvalidEvent, "aggregateID: %s, type: %T", aggregate.GetID(), event)
	}
}

func (s *eventSerializer) DeserializeEvent(event es.Event) (any, error) {
	switch event.GetEventType() {
	case events.BankAccountCreatedEventTypeV1:
		return deserializeEvent(event, new(events.BankAccountCreatedEventV1))
	case events.BalancedDepositedEventTypeV1:
		return deserializeEvent(event, new(events.BalanceDepositedEventV1))
	case events.BalanceWithdrawedEventTypeV1:
		return deserializeEvent(event, new(events.BalanceWithdrawedEventV1))
	default:
		return nil, errors.Wrapf(ErrInvalidEvent, "type: %s", event.GetEventType())
	}

}

func deserializeEvent(event es.Event, targetEvent any) (any, error) {
	if err := event.GetJsonData(&targetEvent); err != nil {
		return nil, errors.Wrapf(err, "event.GetJsonData type: %s", event.GetEventType())
	}
	return targetEvent, nil
}
