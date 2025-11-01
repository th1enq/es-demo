-- Migration script for Event Store
-- This script is idempotent and can be run multiple times safely

-- Create schema for Event Store if not exists
CREATE SCHEMA IF NOT EXISTS microservices;

-- Create events table if not exists
CREATE TABLE IF NOT EXISTS microservices.events (
    event_id BIGSERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    data JSONB NOT NULL,
    version BIGINT NOT NULL,
    metadata JSONB,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(aggregate_id, version)
);

-- Create indexes for faster queries
CREATE INDEX IF NOT EXISTS idx_events_aggregate_id ON microservices.events(aggregate_id);
CREATE INDEX IF NOT EXISTS idx_events_aggregate_type ON microservices.events(aggregate_type);
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON microservices.events(timestamp);

-- Create snapshots table if not exists
CREATE TABLE IF NOT EXISTS microservices.snapshots (
    aggregate_id UUID PRIMARY KEY,
    aggregate_type VARCHAR(255) NOT NULL,
    data JSONB NOT NULL,
    version BIGINT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for snapshots
CREATE INDEX IF NOT EXISTS idx_snapshots_aggregate_type ON microservices.snapshots(aggregate_type);
CREATE INDEX IF NOT EXISTS idx_snapshots_timestamp ON microservices.snapshots(timestamp);

-- Grant permissions (adjust user as needed)
GRANT ALL PRIVILEGES ON SCHEMA microservices TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA microservices TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA microservices TO postgres;
