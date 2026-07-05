package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"event-bus.go/events"
)

type fakeEventBus struct {
	addedEventTypes  []string
	removedEventType string
	publishedEvents  []events.Event
	eventTypes       []string
	subscribeCh      chan events.Event
	subscribeErr     error
	publishErr       error
}

func (f *fakeEventBus) AddEventType(eventType string) error {
	f.addedEventTypes = append(f.addedEventTypes, eventType)
	return nil
}

func (f *fakeEventBus) GetEventTypes() []string {
	return f.eventTypes
}

func (f *fakeEventBus) RemoveEventType(eventType string) {
	f.removedEventType = eventType
}

func (f *fakeEventBus) Publish(evt events.Event) error {
	f.publishedEvents = append(f.publishedEvents, evt)
	return f.publishErr
}

func (f *fakeEventBus) Subscribe(eventTypes []string) (chan events.Event, error) {
	if f.subscribeErr != nil {
		return nil, f.subscribeErr
	}
	if f.subscribeCh == nil {
		f.subscribeCh = make(chan events.Event)
	}
	return f.subscribeCh, nil
}

func (f *fakeEventBus) Unsubscribe(eventTypes []string, busKey any) error {
	return nil
}

func TestHandleHealthCheck(t *testing.T) {
	t.Setenv("APP_VERSION", "1.2.3")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	server := &Server{}
	server.HandleHealthCheck(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var response StatusResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid JSON response, got error: %v", err)
	}

	if response.Status != "ok" || response.Version != "1.2.3" {
		t.Fatalf("expected health response status ok and version 1.2.3, got %+v", response)
	}
}

func TestHandleAddEvent(t *testing.T) {
	bus := &fakeEventBus{}
	server := &Server{EventBus: bus}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(`{"event-type":"user.created"}`))

	server.HandleAddEvent(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !reflect.DeepEqual(bus.addedEventTypes, []string{"user.created"}) {
		t.Fatalf("expected added event types %v, got %v", []string{"user.created"}, bus.addedEventTypes)
	}
}

func TestHandleGetEvents(t *testing.T) {
	bus := &fakeEventBus{eventTypes: []string{"alpha", "beta"}}
	server := &Server{EventBus: bus}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/events", nil)

	server.HandleGetEvents(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var response GetEventsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid JSON response, got error: %v", err)
	}
	if !reflect.DeepEqual(response.EventTypes, []string{"alpha", "beta"}) {
		t.Fatalf("expected event types %v, got %v", []string{"alpha", "beta"}, response.EventTypes)
	}
}

func TestHandleDeleteEvent(t *testing.T) {
	bus := &fakeEventBus{}
	server := &Server{EventBus: bus}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/events/user.deleted", nil)
	request.SetPathValue("event", "user.deleted")

	server.HandleDeleteEvent(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if bus.removedEventType != "user.deleted" {
		t.Fatalf("expected event type user.deleted to be removed, got %q", bus.removedEventType)
	}
}

func TestHandlePublishEventSetsTimestampAndPublishes(t *testing.T) {
	bus := &fakeEventBus{}
	server := &Server{EventBus: bus}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/events/user.created", strings.NewReader(`{"event-type":"user.created","data":{"id":1}}`))
	request.SetPathValue("event", "user.created")

	server.HandlePublishEvent(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, recorder.Code)
	}
	if len(bus.publishedEvents) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(bus.publishedEvents))
	}

	published := bus.publishedEvents[0]
	if published.EventType != "user.created" {
		t.Fatalf("expected event type user.created, got %q", published.EventType)
	}
	if published.EventTime == nil {
		t.Fatal("expected publish handler to set an event timestamp")
	}
	if published.EventData["id"] != float64(1) {
		t.Fatalf("expected event data id 1, got %v", published.EventData["id"])
	}
}

func TestHandleEventStreamStreamsSSE(t *testing.T) {
	bus := &fakeEventBus{subscribeCh: make(chan events.Event, 2)}
	server := &Server{EventBus: bus}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/event-stream?event=alpha&event=beta", nil)

	go func() {
		bus.subscribeCh <- events.Event{EventType: "alpha", EventData: map[string]any{"id": 2}}
		bus.subscribeCh <- events.Event{EventType: "_close"}
	}()

	server.HandleEventStream(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "event: alpha") {
		t.Fatalf("expected SSE event output, got %q", body)
	}
	if !strings.Contains(body, "\"id\":2") {
		t.Fatalf("expected SSE payload to include event data, got %q", body)
	}
}
