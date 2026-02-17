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

type BatchPayload struct {
	Requests []BatchRequest `json:"requests"`
}

type BatchResponse struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body,omitempty"`
}

type Batch struct {
	c        *Client
	requests []BatchRequest
}

func NewBatch(params ...*Client) *Batch {
	if len(params) > 0 && params[0] != nil {
		return &Batch{c: params[0]}
	}
	return &Batch{c: mustDefault()}
}

func (b *Batch) Reset() { b.requests = b.requests[:0] }

type BatchCollection struct {
	b    *Batch
	name string
}

func (b *Batch) Collection(name string) *BatchCollection { return &BatchCollection{b: b, name: name} }

func (bc *BatchCollection) Create(body any, params ...string) {
	bc.b.add(http.MethodPost, "/api/collections/"+url.PathEscape(bc.name)+"/records", body, params...)
}
func (bc *BatchCollection) Update(id string, body any, params ...string) {
	bc.b.add(http.MethodPatch, "/api/collections/"+url.PathEscape(bc.name)+"/records/"+url.PathEscape(id), body, params...)
}
func (bc *BatchCollection) Upsert(body any, params ...string) {
	bc.b.add(http.MethodPut, "/api/collections/"+url.PathEscape(bc.name)+"/records", body, params...)
}
func (bc *BatchCollection) Delete(id string, params ...string) {
	bc.b.add(http.MethodDelete, "/api/collections/"+url.PathEscape(bc.name)+"/records/"+url.PathEscape(id), nil, params...)
}

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

func (b *Batch) Send(params ...string) ([]BatchResponse, error) {
	var out []BatchResponse
	payload := BatchPayload{Requests: b.requests}
	err := b.c.doJSON(http.MethodPost, "/api/batch", optParam(params), payload, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
