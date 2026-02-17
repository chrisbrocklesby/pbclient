package pbclient

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// Batch endpoint runs all sub-requests under the same auth context as the outer /api/batch request.
// Per-request Authorization is not supported by PocketBase.
type BatchRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    any               `json:"body,omitempty"`
}

// BatchPayload is the request body sent to /api/batch.
type BatchPayload struct {
	Requests []BatchRequest `json:"requests"`
}

// BatchResponse represents one sub-response returned by /api/batch.
type BatchResponse struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body,omitempty"`
}

// Batch accumulates requests that are sent in a single /api/batch call.
type Batch struct {
	c        *Client
	requests []BatchRequest
}

// NewBatch creates a new batch bound to the provided client or the default client.
func NewBatch(params ...*Client) *Batch {
	if len(params) > 0 && params[0] != nil {
		return &Batch{c: params[0]}
	}
	return &Batch{c: mustDefault()}
}

// Reset clears queued requests.
func (b *Batch) Reset() { b.requests = b.requests[:0] }

// BatchCollection provides collection helpers that enqueue batch requests.
type BatchCollection struct {
	b    *Batch
	name string
}

// Collection returns batch helpers for a specific collection name.
func (b *Batch) Collection(name string) *BatchCollection { return &BatchCollection{b: b, name: name} }

// Create enqueues a record create operation.
func (bc *BatchCollection) Create(body any, params ...string) {
	bc.b.add(http.MethodPost, "/api/collections/"+url.PathEscape(bc.name)+"/records", body, params...)
}

// Update enqueues a record update operation.
func (bc *BatchCollection) Update(id string, body any, params ...string) {
	bc.b.add(http.MethodPatch, "/api/collections/"+url.PathEscape(bc.name)+"/records/"+url.PathEscape(id), body, params...)
}

// Upsert enqueues a collection upsert operation.
func (bc *BatchCollection) Upsert(body any, params ...string) {
	bc.b.add(http.MethodPut, "/api/collections/"+url.PathEscape(bc.name)+"/records", body, params...)
}

// Delete enqueues a record delete operation.
func (bc *BatchCollection) Delete(id string, params ...string) {
	bc.b.add(http.MethodDelete, "/api/collections/"+url.PathEscape(bc.name)+"/records/"+url.PathEscape(id), nil, params...)
}

// Raw enqueues an arbitrary request for /api/batch.
func (b *Batch) Raw(method, urlPath string, body any) {
	b.requests = append(b.requests, BatchRequest{Method: method, URL: urlPath, Body: body})
}

func (b *Batch) add(method, basePath string, body any, params ...string) {
	u := basePath
	q := optParam(params)
	if q != "" {
		u += "?" + q
	}
	b.requests = append(b.requests, BatchRequest{Method: method, URL: u, Body: body})
}

// Send posts all queued requests to /api/batch and returns sub-responses.
func (b *Batch) Send(params ...string) ([]BatchResponse, error) {
	var out []BatchResponse
	payload := BatchPayload{Requests: b.requests}
	err := b.c.doJSON(http.MethodPost, "/api/batch", optParam(params), payload, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
