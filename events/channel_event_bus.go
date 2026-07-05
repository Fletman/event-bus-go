package events

import (
	"fmt"
	"slices"
	"sync"
)

type ChannelEventBus struct {
	event_listeners map[string]map[chan Event]struct{}
	event_sync      *sync.RWMutex
}

func (eb *ChannelEventBus) AddEventType(event_type string) error {
	if event_type[0] == '_' {
		return fmt.Errorf("Event name '%s' cannot begin with underscore", event_type)
	}
	eb.event_sync.Lock()
	defer eb.event_sync.Unlock()
	eb.event_listeners[event_type] = make(map[chan Event]struct{})
	return nil
}

func (eb *ChannelEventBus) GetEventTypes() []string {
	eb.event_sync.RLock()
	defer eb.event_sync.RUnlock()

	event_types := make([]string, len(eb.event_listeners))
	i := 0
	for event_type := range eb.event_listeners {
		event_types[i] = event_type
		i++
	}

	slices.Sort(event_types)

	return event_types
}

func (eb *ChannelEventBus) RemoveEventType(event_type string) {
	close_event := Event{EventType: "_close"}
	for event_channel := range eb.event_listeners[event_type] {
		event_channel <- close_event
	}

	eb.event_sync.Lock()
	defer eb.event_sync.Unlock()
	delete(eb.event_listeners, event_type)
}

func (eb *ChannelEventBus) Publish(event Event) error {
	eb.event_sync.RLock()
	event_channels, ok := eb.event_listeners[event.EventType]
	eb.event_sync.RUnlock()
	if !ok {
		return fmt.Errorf("Could not find event type: %s", event.EventType)
	}
	for ec := range event_channels {
		ec <- event
	}
	return nil
}

func (eb *ChannelEventBus) Subscribe(event_types []string) (chan Event, error) {
	event_channel := make(chan Event)
	eb.event_sync.Lock()
	for _, event_type := range event_types {
		eb.event_listeners[event_type][event_channel] = struct{}{}
	}
	eb.event_sync.Unlock()
	return event_channel, nil
}

func (eb *ChannelEventBus) Unsubscribe(event_types []string, bus_key any) error {
	defer close(bus_key.(chan Event))
	eb.event_sync.Lock()
	for _, event_type := range event_types {
		delete(eb.event_listeners[event_type], bus_key.(chan Event))
	}
	eb.event_sync.Unlock()
	return nil
}

func NewChannelEventBus() *ChannelEventBus {
	return &ChannelEventBus{
		event_listeners: make(map[string]map[chan Event]struct{}),
		event_sync:      new(sync.RWMutex),
	}
}
