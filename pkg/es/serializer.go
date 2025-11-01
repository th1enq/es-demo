package es

type Serializer interface {
	SerializeEvent(aggregate Aggregate, event any) (Event, error)
	DeserializeEvent(event Event) (any, error)
}
