package es

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	eventsCapacity = 10
)

type pgEventStore struct {
	logger     *zap.Logger
	cfg        Config
	db         *pgxpool.Pool
	eventBus   EventsBus
	serializer Serializer
}

func NewPgEventStore(
	cfg Config,
	db *pgxpool.Pool,
	serializer Serializer,
	logger *zap.Logger,
	eventBus EventsBus,
) *pgEventStore {
	return &pgEventStore{
		cfg:        cfg,
		db:         db,
		serializer: serializer,
		logger:     logger,
		eventBus:   eventBus,
	}
}

// SaveEvents save aggregate uncommitted events as one batch and process with event bus using transaction
func (p *pgEventStore) SaveEvents(ctx context.Context, events []Event) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		p.logger.Error("(Save Events) db.Begin error", zap.Error(err))
		return errors.Wrap(err, "db.Begin")
	}

	if err := p.handleConcurrency(ctx, tx, events); err != nil {
		return RollBackTx(ctx, tx, err)
	}

	// If aggregate changes has single event save it
	if len(events) == 1 {
		_, err := tx.Exec(
			ctx,
			saveEventQuery,
			events[0].GetAggregateID(),
			events[0].GetAggregateType(),
			events[0].GetEventType(),
			events[0].GetData(),
			events[0].GetVersion(),
			events[0].GetMetadata(),
		)
		if err != nil {
			p.logger.Error("(Save Events) tx.Exec error", zap.Error(err))
			return RollBackTx(ctx, tx, err)
		}

		p.logger.Debug("(Save Events) result",
			zap.String("aggregate_id", events[0].GetAggregateID()),
			zap.Uint64("event_version", events[0].GetVersion()),
		)

		return tx.Commit(ctx)
	}

	batch := &pgx.Batch{}
	for _, event := range events {
		batch.Queue(
			saveEventQuery,
			event.GetAggregateID(),
			event.GetAggregateType(),
			event.GetEventType(),
			event.GetData(),
			event.GetVersion(),
			event.GetMetadata(),
		)
	}

	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		p.logger.Error("(Save Events) tx.SendBatch error", zap.Error(err))
		return RollBackTx(ctx, tx, err)
	}

	return tx.Commit(ctx)
}

// LoadEvents load aggregate events by id
func (p *pgEventStore) LoadEvents(ctx context.Context, aggregateID string) ([]Event, error) {
	rows, err := p.db.Query(ctx, getEventsQuery, aggregateID)
	if err != nil {
		p.logger.Error("(Load Events) db.Query error", zap.Error(err))
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	events := make([]Event, 0, eventsCapacity)

	for rows.Next() {
		var event Event
		if err := rows.Scan(
			&event.EventID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Data,
			&event.Version,
			&event.Timestamp,
			&event.Metadata,
		); err != nil {
			p.logger.Error("(Load Events) rows.Scan error", zap.Error(err))
			return nil, errors.Wrap(err, "rows.Scan")
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		p.logger.Error("(Load Events) rows.Err error", zap.Error(err))
		return nil, errors.Wrap(err, "rows.Err")
	}

	return events, nil
}

// LoadEvents load aggregate events by id
func (p *pgEventStore) loadEvents(ctx context.Context, aggregate Aggregate) error {
	rows, err := p.db.Query(ctx, getEventsQuery, aggregate.GetID())
	if err != nil {
		p.logger.Error("(Load Events) db.Query error", zap.Error(err))
		return errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	for rows.Next() {
		var event Event

		if err := rows.Scan(
			&event.EventID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Data,
			&event.Version,
			&event.Timestamp,
			&event.Metadata,
		); err != nil {
			return errors.Wrap(err, "rows.Scan")
		}

		deserializedEvent, err := p.serializer.DeserializeEvent(event)
		if err != nil {
			p.logger.Error("(Load Events) serializer.DeserializeEvent error", zap.Error(err))
			return errors.Wrap(err, "serializer.DeserializeEvent")
		}

		if err := aggregate.RaiseEvent(deserializedEvent); err != nil {
			p.logger.Error("(Load Events) RaiseEvent error", zap.Error(err))
			return errors.Wrap(err, "RaiseEvent")
		}
	}

	if err := rows.Err(); err != nil {
		p.logger.Error("(Load Events) rows.Err error", zap.Error(err))
		return errors.Wrap(err, "rows.Err")
	}

	return nil
}

// Exists check for exists aggregate by id
func (p *pgEventStore) Exists(ctx context.Context, aggregateID string) (bool, error) {
	var id string
	if err := p.db.QueryRow(ctx, getEventQuery, aggregateID).Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		p.logger.Error("(Exists) db.QueryRow error", zap.Error(err))
		return false, errors.Wrap(err, "db.QueryRow")
	}
	p.logger.Debug("(Exists Aggregate)", zap.String("id", id))

	return true, nil
}

func (p *pgEventStore) loadEventsByVersion(ctx context.Context, aggregateID string, versionFrom uint64) ([]Event, error) {
	rows, err := p.db.Query(ctx, getEventsByVersionQuery, aggregateID, versionFrom)
	if err != nil {
		p.logger.Error("(Load Events) db.Query error", zap.Error(err))
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	events := make([]Event, 0, p.cfg.SnapshotFrequency)

	for rows.Next() {
		var event Event

		if err := rows.Scan(
			&event.EventID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Data,
			&event.Version,
			&event.Timestamp,
			&event.Metadata,
		); err != nil {
			p.logger.Error("(Load Events) rows.Scan error", zap.Error(err))
			return nil, errors.Wrap(err, "rows.Scan")
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		p.logger.Error("(Load Events) rows.Err error", zap.Error(err))
		return nil, errors.Wrap(err, "rows.Err")
	}

	return events, nil
}

func (p *pgEventStore) loadAggregateEventsByVersion(ctx context.Context, aggregate Aggregate) error {
	rows, err := p.db.Query(ctx, getEventsByVersionQuery, aggregate.GetID(), aggregate.GetVersion())
	if err != nil {
		p.logger.Error("(Load Events By Version) db.Query error", zap.Error(err))
		return errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	for rows.Next() {
		var event Event

		if err := rows.Scan(
			&event.EventID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Data,
			&event.Version,
			&event.Timestamp,
			&event.Metadata,
		); err != nil {
			p.logger.Error("(Load Events By Version) rows.Scan error", zap.Error(err))
			return errors.Wrap(err, "rows.Scan")
		}

		deserializedEvent, err := p.serializer.DeserializeEvent(event)
		if err != nil {
			p.logger.Error("(Load Events By Version) serializer.DeserializeEvent error", zap.Error(err))
			return errors.Wrap(err, "serializer.DeserializeEvent")
		}

		if err := aggregate.RaiseEvent(deserializedEvent); err != nil {
			p.logger.Error("(Load Events By Version) RaiseEvent error", zap.Error(err))
			return errors.Wrap(err, "RaiseEvent")
		}
	}

	if err := rows.Err(); err != nil {
		p.logger.Error("(Load Events By Version) rows.Err error", zap.Error(err))
		return errors.Wrap(err, "rows.Err")
	}

	return nil
}

func (p *pgEventStore) loadEventsByVersionTx(ctx context.Context, tx pgx.Tx, aggregateID string, versionFrom int64) ([]Event, error) {
	rows, err := tx.Query(ctx, getEventsByVersionQuery, aggregateID, versionFrom)
	if err != nil {
		p.logger.Error("(Load Events) db.Query error", zap.Error(err))
		return nil, errors.Wrap(err, "tx.Query")
	}
	defer rows.Close()

	events := make([]Event, 0, p.cfg.SnapshotFrequency)

	for rows.Next() {
		var event Event

		if err := rows.Scan(
			&event.EventID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Data,
			&event.Version,
			&event.Timestamp,
			&event.Metadata,
		); err != nil {
			p.logger.Error("(Load Events) rows.Scan error", zap.Error(err))
			return nil, errors.Wrap(err, "rows.Scan")
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		p.logger.Error("(Load Events) rows.Err error", zap.Error(err))
		return nil, errors.Wrap(err, "rows.Err")
	}

	return events, nil
}

func (p *pgEventStore) handleConcurrency(ctx context.Context, tx pgx.Tx, events []Event) error {
	result, err := tx.Exec(ctx, handleConcurrentWriteQuery, events[0].GetAggregateID())
	if err != nil {
		p.logger.Error("(Handle Concurrency) tx.Exec error", zap.Error(err))
		return errors.Wrap(err, "tx.Exec")
	}

	p.logger.Debug("(Handle Concurrency) success", zap.String("result", result.String()))

	return nil
}

func (p *pgEventStore) saveEventsTx(ctx context.Context, tx pgx.Tx, events []Event) error {
	if err := p.handleConcurrency(ctx, tx, events); err != nil {
		return err
	}

	if len(events) == 1 {
		result, err := tx.Exec(
			ctx,
			saveEventQuery,
			events[0].GetAggregateID(),
			events[0].GetAggregateType(),
			events[0].GetEventType(),
			events[0].GetData(),
			events[0].GetVersion(),
			events[0].GetMetadata(),
		)
		if err != nil {
			p.logger.Error("(Save Events) tx.Exec error", zap.Error(err))
			return errors.Wrap(err, "tx.Exec")
		}

		p.logger.Debug("(saveEventsTx)",
			zap.String("result", result.String()),
			zap.String("aggregate_id", events[0].GetAggregateID()),
			zap.Uint64("event_version", events[0].GetVersion()),
		)

		return nil
	}

	batch := &pgx.Batch{}
	for _, event := range events {
		batch.Queue(
			saveEventQuery,
			event.GetAggregateID(),
			event.GetAggregateType(),
			event.GetEventType(),
			event.GetData(),
			event.GetVersion(),
			event.GetMetadata(),
		)
	}

	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		p.logger.Error("(Save Events) tx.SendBatch error", zap.Error(err))
		return errors.Wrap(err, "tx.SendBatch")
	}

	return nil
}

func (p *pgEventStore) saveSnapshotTx(ctx context.Context, tx pgx.Tx, aggregate Aggregate) error {
	snapshot, err := NewSnapshotFromAggregate(aggregate)
	if err != nil {
		p.logger.Error("(Save Snapshot) NewSnapshotFromAggregate error", zap.Error(err))
		return err
	}

	_, err = tx.Exec(ctx, saveSnapshotQuery, snapshot.ID, snapshot.Type, snapshot.State, snapshot.Version)
	if err != nil {
		p.logger.Error("(Save Snapshot) tx.Exec error", zap.Error(err))
		return errors.Wrap(err, "tx.Exec")
	}

	p.logger.Debug("(Save Snapshot) success", zap.String("snapshot", snapshot.String()))

	return nil
}

func (p *pgEventStore) processEvents(ctx context.Context, events []Event) error {
	p.logger.Info("Processing events", zap.Int("count", len(events)))
	return p.eventBus.ProcessEvents(ctx, events)
}

func RollBackTx(ctx context.Context, tx pgx.Tx, err error) error {
	if err := tx.Rollback(ctx); err != nil {
		return errors.Wrap(err, "tx.Rollback")
	}
	return err
}
