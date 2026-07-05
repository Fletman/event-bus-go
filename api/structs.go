package api

type RequestLog struct {
	URI     string              `json:"uri"`
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
	Host    string              `json:"host"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type StatusResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type AddEventRequest struct {
	EventType string `json:"event-type"`
}

type GetEventsResponse struct {
	EventTypes []string `json:"event-types"`
}
