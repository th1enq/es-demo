package es

const (
	saveEventQuery = `INSERT INTO microservices.events as e (aggregate_id, aggregate_type, event_type, data, version, metadata, timestamp)
	VALUES ($1, $2, $3, $4, $5, $6, now())`

	getEventsQuery = `SELECT event_id, aggregate_id, aggregate_type, event_type, data, version, timestamp, metadata 
	FROM microservices.events e WHERE aggregate_id = $1 ORDER BY version ASC`

	getEventQuery = `SELECT aggregate_id FROM microservices.events e WHERE aggregate_id = $1`

	getEventsByVersionQuery = `SELECT event_id, aggregate_id, aggregate_type, event_type, data, version, timestamp, metadata 
	FROM microservices.events e WHERE aggregate_id = $1 AND version > $2 ORDER BY version ASC`

	getEventsByVersionRangeQuery = `SELECT event_id, aggregate_id, aggregate_type, event_type, data, version, timestamp, metadata 
	FROM microservices.events e WHERE aggregate_id = $1 AND version BETWEEN $2 AND $3 ORDER BY version ASC`

	getAllEventsQuery = `SELECT event_id, aggregate_id, aggregate_type, event_type, data, version, timestamp, metadata 
	FROM microservices.events e ORDER BY timestamp ASC, version ASC`

	saveSnapshotQuery = `INSERT INTO microservices.snapshots (aggregate_id, aggregate_type, data, version, timestamp)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (aggregate_id, version)
		DO UPDATE SET data = EXCLUDED.data, timestamp = now()`

	getSnapshotQuery = `SELECT aggregate_id, aggregate_type, data, version FROM microservices.snapshots s WHERE aggregate_id = $1 ORDER BY version DESC LIMIT 1`

	getSnapshotByVersionQuery = `SELECT aggregate_id, aggregate_type, data, version FROM microservices.snapshots WHERE aggregate_id = $1 AND version = $2`

	handleConcurrentWriteQuery = `SELECT aggregate_id FROM microservices.events e WHERE e.aggregate_id = $1 LIMIT 1 FOR UPDATE`
)
