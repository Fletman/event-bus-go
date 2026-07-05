package api

import (
	"fmt"
	"log"
	"net/http"

	"event-bus.go/events"
)

type EventManager struct {
}

type TlsConfig struct {
	CertPath string
	KeyPath  string
}

type Server struct {
	EventBus events.EventBus
}

func NewServer(event_bus_config string) (server *Server, err error) {
	var event_bus events.EventBus
	switch event_bus_config {
	case "channel":
		event_bus = events.NewChannelEventBus()
	default:
		err = fmt.Errorf("Unsupported event bus option: %s", event_bus_config)
		return
	}

	server = &Server{
		EventBus: event_bus,
	}
	return
}

func (s *Server) StartServer(port string, tls *TlsConfig) (err error) {
	log.Printf("Starting server on port: %s\n", port)

	router := http.NewServeMux()
	s.AddHandlers(router)

	if tls == nil {
		err = http.ListenAndServe(":"+port, router)
	} else {
		err = http.ListenAndServeTLS(
			":"+port,
			tls.CertPath,
			tls.KeyPath,
			router,
		)
	}
	return
}
