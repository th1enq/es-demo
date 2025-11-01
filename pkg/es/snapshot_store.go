package es

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// SaveSnapshot save es.Aggregate snapshot
func (p *pgEventStore) SaveSnapshot(ctx context.Context, aggregate Aggregate) error {
	p.logger.Info("Saving snapshot", zap.String("aggregateID", aggregate.String()))
	snapshot, err := NewSnapshotFromAggregate(aggregate)
	if err != nil {
		p.logger.Error("(Save Snapshot) NewSnapshotFromAggregate error", zap.Error(err))
		return errors.Wrap(err, "NewSnapshotFromAggregate")
	}

	_, err = p.db.Exec(ctx, saveSnapshotQuery, snapshot.ID, snapshot.Type, snapshot.State, snapshot.Version)
	if err != nil {
		p.logger.Error("(Save Snapshot) db.Exec error", zap.Error(err))
		return errors.Wrap(err, "db.Exec")
	}
	p.logger.Info("Snapshot saved successfully", zap.String("snapshot", snapshot.String()))
	return nil
}

// GetSnapshot load es.Aggregate snapshot
func (p *pgEventStore) GetSnapshot(ctx context.Context, id string) (*Snapshot, error) {
	p.logger.Info("Get Snapshot", zap.String("aggregateID", id))
	var snapshot Snapshot
	if err := p.db.QueryRow(ctx, getSnapshotQuery, id).Scan(&snapshot.ID, &snapshot.Type, &snapshot.State, &snapshot.Version); err != nil {
		p.logger.Error("(Get Snapshot) db.QueryRow error", zap.Error(err))
		return nil, errors.Wrap(err, "db.QueryRow")
	}
	p.logger.Info("Get Snapshot successfully", zap.String("snapshot", snapshot.String()))

	return &snapshot, nil
}

func (p *pgEventStore) GetSnapshotByVersion(ctx context.Context, id string, version uint64) (*Snapshot, error) {
	p.logger.Info("Get Snapshot By Version", zap.String("aggregateID", id), zap.Uint64("version", version))
	var snapshot Snapshot
	if err := p.db.QueryRow(ctx, getSnapshotByVersionQuery, id, version).Scan(&snapshot.ID, &snapshot.Type, &snapshot.State, &snapshot.Version); err != nil {
		p.logger.Error("(Get Snapshot By Version) db.QueryRow error", zap.Error(err))
		return nil, errors.Wrap(err, "db.QueryRow")
	}
	p.logger.Info("Get Snapshot By Version successfully", zap.String("snapshot", snapshot.String()))

	return &snapshot, nil
}
