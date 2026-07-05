package api

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
