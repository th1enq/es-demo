package es

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/th1enq/es-demo/pkg/es/serializer"
	"go.uber.org/zap"
)

// Load es.Aggregate events using snapshots with given frequency
func (p *pgEventStore) Load(ctx context.Context, aggregate Aggregate) error {
	p.logger.Info("Loading aggregate", zap.String("aggregateID", aggregate.String()))
	snapshot, err := p.GetSnapshot(ctx, aggregate.GetID())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if snapshot != nil {
		if err := serializer.Unmarshal(snapshot.State, aggregate); err != nil {
			p.logger.Error("(Load) serializer.Unmarshal failed", zap.Error(err))
			return errors.Wrap(err, "json.Unmarshal")
		}

		err := p.loadAggregateEventsByVersion(ctx, aggregate)
		if err != nil {
			return err
		}
		p.logger.Debug("Load Aggregate By Version", zap.String("aggregate", aggregate.String()))
		return nil
	}

	err = p.loadEvents(ctx, aggregate)
	if err != nil {
		return err
	}

	p.logger.Debug("Load Aggregate successfully", zap.String("aggregate", aggregate.String()))
	return nil
}

func (p *pgEventStore) LoadByVersion(ctx context.Context, aggregate Aggregate, version uint64) error {
	p.logger.Info("Loading aggregate", zap.String("aggregateID", aggregate.String()), zap.Uint64("version", version))

	snapshotVersion := version / p.cfg.SnapshotFrequency * p.cfg.SnapshotFrequency

	snapshot, err := p.GetSnapshotByVersion(ctx, aggregate.GetID(), snapshotVersion)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if snapshot != nil {
		if err := serializer.Unmarshal(snapshot.State, aggregate); err != nil {
			p.logger.Error("(Load) serializer.Unmarshal failed", zap.Error(err))
			return errors.Wrap(err, "json.Unmarshal")
		}

		if snapshot.Version < version {
			err := p.loadAggregateEventsByVersionRange(ctx, aggregate, snapshot.Version+1, version)
			if err != nil {
				return err
			}
		}
		p.logger.Debug("Load Aggregate By Version", zap.String("aggregate", aggregate.String()))
		return nil
	}

	err = p.loadEventsIntoVersion(ctx, aggregate, version)
	if err != nil {
		return err
	}

	p.logger.Debug("Load Aggregate successfully", zap.String("aggregate", aggregate.String()))
	return nil
}

// Save es.Aggregate events using snapshots with given frequency
func (p *pgEventStore) Save(ctx context.Context, aggregate Aggregate) (err error) {
	if len(aggregate.GetChanges()) == 0 {
		p.logger.Debug("Save Aggregate: no changes to save", zap.String("aggregate", aggregate.String()))
		return nil
	}

	p.logger.Info("Save Aggregate", zap.String("aggregate", aggregate.String()))

	tx, err := p.db.Begin(ctx)
	if err != nil {
		p.logger.Error("Failed to begin transaction", zap.Error(err))
		return errors.Wrap(err, "db.Begin")
	}

	defer func() {
		if tx != nil {
			if txErr := tx.Rollback(ctx); txErr != nil && !errors.Is(txErr, pgx.ErrTxClosed) {
				err = txErr
				return
			}
		}
	}()

	changes := aggregate.GetChanges()
	events := make([]Event, 0, len(changes))

	for i := range changes {
		event, err := p.serializer.SerializeEvent(aggregate, changes[i])
		if err != nil {
			p.logger.Error("Failed to serialize event", zap.Error(err))
			return errors.Wrap(err, "serializer.SerializeEvent")
		}
		events = append(events, event)
	}

	if err := p.saveEventsTx(ctx, tx, events); err != nil {
		p.logger.Error("Failed to save events", zap.Error(err))
		return errors.Wrap(err, "saveEventsTx")
	}

	if aggregate.GetVersion()%p.cfg.SnapshotFrequency == 0 {
		aggregate.ToSnapshot()
		if err := p.saveSnapshotTx(ctx, tx, aggregate); err != nil {
			return errors.Wrap(err, "saveSnapshotTx")
		}
	}

	if err := p.processEvents(ctx, events); err != nil {
		p.logger.Error("Failed to process events", zap.Error(err))
		return errors.Wrap(err, "processEvents")
	}

	p.logger.Info("Save Aggregate successfully", zap.String("aggregate", aggregate.String()))
	return tx.Commit(ctx)
}
