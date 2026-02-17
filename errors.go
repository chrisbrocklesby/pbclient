package pbclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// APIError captures PocketBase error payloads with HTTP status and message.
type APIError struct {
	Status  int
	Body    []byte
	Message string
	Data    map[string]any
}

func (e *APIError) Error() string {
	msg := strings.TrimSpace(e.Message)
	if msg == "" {
		msg = strings.TrimSpace(string(e.Body))
	}
	if msg == "" {
		msg = http.StatusText(e.Status)
	}
	return fmt.Sprintf("pbclient: http %d: %s", e.Status, msg)
}

func parseAPIError(status int, body []byte) *APIError {
	var parsed struct {
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	}
	_ = json.Unmarshal(body, &parsed)
	return &APIError{Status: status, Body: body, Message: parsed.Message, Data: parsed.Data}
}
