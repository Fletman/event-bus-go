package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"event-bus.go/events"
)

func parseBody(ptr any, b io.ReadCloser) error {
	bytes, err := io.ReadAll(b)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, ptr)
}

func logRequest(r *http.Request) {
	log.Println(r)
}

func eventToSSE(evt events.Event) (sse string, err error) {
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}
	sse = fmt.Sprintf("event: %s\ndata: %s", evt.EventType, string(data))
	return
}

func sseStream(w http.ResponseWriter) {
	headers := [][]string{
		{"Content-Type", "text/event-stream"},
		{"Cache-Control", "no-cache"},
		{"Connection", "keep-alive"},
	}
	for _, h := range headers {
		w.Header().Set(h[0], h[1])
	}
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func (*Server) HandleHealthCheck(w http.ResponseWriter, _ *http.Request) {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "0.0"
	}
	status := StatusResponse{
		Status:  "ok",
		Version: version,
	}
	if err := Ok(w, status); err != nil {
		log.Println(err)
	}
}

func (s *Server) HandleAddEvent(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	req := AddEventRequest{}
	if err := parseBody(&req, r.Body); err != nil {
		log.Println(err)
		if err := BadRequest(w, "Could not parse JSON body"); err != nil {
			log.Println(err)
		}
	}
	if req.EventType == "" {
		BadRequest(w, fmt.Sprintf("Invalid event name: %s", req.EventType))
	}

	if err := s.EventBus.AddEventType(req.EventType); err != nil {
		log.Println(err)
		BadRequest(w, err.Error())
		return
	}

	res := MessageResponse{
		Message: fmt.Sprintf("Added event type: %s", req.EventType),
	}
	if err := Ok(w, res); err != nil {
		log.Println(err)
	}
}

func (s *Server) HandleGetEvents(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	res := GetEventsResponse{EventTypes: s.EventBus.GetEventTypes()}
	if err := Ok(w, res); err != nil {
		log.Println(err)
	}
}

func (s *Server) HandleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	event_type := r.PathValue("event")
	if event_type == "" {
		if err := BadRequest(w, "No event type specified"); err != nil {
			log.Println(err)
		}
	} else {
		s.EventBus.RemoveEventType(event_type)
		res := MessageResponse{
			Message: fmt.Sprintf("Successfully deleted event: %s", event_type),
		}
		if err := Ok(w, res); err != nil {
			log.Println(err)
		}
	}
}

func (s *Server) HandlePublishEvent(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	event_type := r.PathValue("event")
	if event_type == "" {
		if err := BadRequest(w, "No event type specified"); err != nil {
			log.Println(err)
		}
		return
	}
	evt := events.Event{}
	if err := parseBody(&evt, r.Body); err != nil {
		log.Println(err)
		if err := BadRequest(w, "Could not parse JSON body"); err != nil {
			log.Println(err)
		}
		return
	}

	evt.EventType = event_type
	if evt.EventTime == nil {
		timestamp := time.Now().Unix()
		evt.EventTime = &timestamp
	}
	if err := s.EventBus.Publish(evt); err != nil {
		log.Println(err)
		if strings.Contains(err.Error(), "Could not find") {
			NotFound(w, err.Error())
		} else {
			InternalServerError(w)
		}
		return
	}

	res := MessageResponse{
		Message: fmt.Sprintf("Message published at timestamp: %d", *evt.EventTime),
	}
	Accepted(w, res)
}

func (s *Server) HandleEventStream(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	event_types := r.URL.Query()["event"]
	if len(event_types) == 0 {
		if err := BadRequest(w, "No event list found"); err != nil {
			log.Println(err)
		}
		return
	}

	event_channel, err := s.EventBus.Subscribe(event_types)
	if err != nil {
		log.Println(err)
		InternalServerError(w)
		return
	}
	defer close(event_channel)

	sseStream(w)
	for evt := <-event_channel; evt.EventType != "_close"; evt = <-event_channel {
		sse, err := eventToSSE(evt)
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Fprint(w, sse)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func (s *Server) AddHandlers(router *http.ServeMux) {
	router.HandleFunc("/health", s.HandleHealthCheck)
	router.HandleFunc("POST /events", s.HandleAddEvent)
	router.HandleFunc("GET /events", s.HandleGetEvents)
	router.HandleFunc("DELETE /events/{event}", s.HandleDeleteEvent)
	router.HandleFunc("POST /events/{event}", s.HandlePublishEvent)
	router.HandleFunc("GET /event-stream", s.HandleEventStream)
}
