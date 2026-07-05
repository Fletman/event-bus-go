package events

type Event struct {
	EventType string         `json:"event-type"`
	EventTime *int64         `json:"event-time"`
	EventData map[string]any `json:"data"`
}

type EventBus interface {
	AddEventType(event_type string) error
	GetEventTypes() []string
	RemoveEventType(event_type string)
	Publish(event Event) error
	Subscribe(event_types []string) (chan Event, error)
	Unsubscribe(event_types []string, bus_key any) error
}
