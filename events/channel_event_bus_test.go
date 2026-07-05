package events

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestAddEventTypeAndGetEventTypes(t *testing.T) {
	eb := NewChannelEventBus()

	if err := eb.AddEventType("alpha"); err != nil {
		t.Fatalf("expected alpha to be added, got error: %v", err)
	}
	if err := eb.AddEventType("beta"); err != nil {
		t.Fatalf("expected beta to be added, got error: %v", err)
	}

	err := eb.AddEventType("_hidden")
	if err == nil || !strings.Contains(err.Error(), "cannot begin with underscore") {
		t.Fatalf("expected underscore validation error, got %v", err)
	}

	got := eb.GetEventTypes()
	want := []string{"alpha", "beta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected event types %v, got %v", want, got)
	}
}

func TestPublishDeliversEventToSubscribers(t *testing.T) {
	eb := NewChannelEventBus()
	if err := eb.AddEventType("user.created"); err != nil {
		t.Fatalf("expected event type registration to succeed, got error: %v", err)
	}

	sub, err := eb.Subscribe([]string{"user.created"})
	if err != nil {
		t.Fatalf("expected subscription to succeed, got error: %v", err)
	}

	event := Event{EventType: "user.created", EventData: map[string]any{"id": 1}}
	publishErrCh := make(chan error, 1)
	go func() {
		publishErrCh <- eb.Publish(event)
	}()

	select {
	case got := <-sub:
		if got.EventType != event.EventType {
			t.Fatalf("expected event type %q, got %q", event.EventType, got.EventType)
		}
		if got.EventData["id"] != event.EventData["id"] {
			t.Fatalf("expected event data %v, got %v", event.EventData, got.EventData)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for published event")
	}

	select {
	case err := <-publishErrCh:
		if err != nil {
			t.Fatalf("expected publish to succeed, got error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for publish completion")
	}
}

func TestPublishReturnsErrorForUnknownEventType(t *testing.T) {
	eb := NewChannelEventBus()

	err := eb.Publish(Event{EventType: "missing"})
	if err == nil || !strings.Contains(err.Error(), "Could not find event type") {
		t.Fatalf("expected unknown event error, got %v", err)
	}
}

func TestUnsubscribeRemovesSubscriberAndRemoveEventTypeNotifies(t *testing.T) {
	eb := NewChannelEventBus()
	if err := eb.AddEventType("user.deleted"); err != nil {
		t.Fatalf("expected event type registration to succeed, got error: %v", err)
	}

	sub, err := eb.Subscribe([]string{"user.deleted"})
	if err != nil {
		t.Fatalf("expected subscription to succeed, got error: %v", err)
	}

	if err := eb.Unsubscribe([]string{"user.deleted"}, sub); err != nil {
		t.Fatalf("expected unsubscription to succeed, got error: %v", err)
	}

	event := Event{EventType: "user.deleted"}
	if err := eb.Publish(event); err != nil {
		t.Fatalf("expected publish to succeed after unsubscribe, got error: %v", err)
	}

	select {
	case _, ok := <-sub:
		if ok {
			t.Fatal("expected subscription channel to be closed after unsubscribe")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for unsubscribe to close the channel")
	}

	// Re-register and ensure removal notifies subscribers with a close event.
	eb = NewChannelEventBus()
	if err := eb.AddEventType("user.deleted"); err != nil {
		t.Fatalf("expected event type registration to succeed, got error: %v", err)
	}

	reopened, err := eb.Subscribe([]string{"user.deleted"})
	if err != nil {
		t.Fatalf("expected second subscription to succeed, got error: %v", err)
	}

	removeDone := make(chan struct{})
	go func() {
		eb.RemoveEventType("user.deleted")
		close(removeDone)
	}()

	select {
	case received := <-reopened:
		if received.EventType != "_close" {
			t.Fatalf("expected close event, got %q", received.EventType)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for RemoveEventType notification")
	}

	select {
	case <-removeDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for RemoveEventType to complete")
	}
}
