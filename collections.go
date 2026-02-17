package pbclient

import (
	"fmt"
	"net/http"
	"net/url"
)

type ListResult[T any] struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
	Items      []T `json:"items"`
}

// Coll provides typed CRUD helpers for a PocketBase collection.
type Coll[T any] struct {
	c    *Client
	name string
}

// Collection returns a typed collection. If a client is provided, it is used; otherwise the package default client is used.
func Collection[T any](name string, client ...*Client) *Coll[T] {
	c := mustDefault()
	if len(client) > 0 && client[0] != nil {
		c = client[0]
	}
	return &Coll[T]{c: c, name: name}
}

// Create inserts a record in the collection.
func (col *Coll[T]) Create(record any, params ...string) (T, error) {
	var out T
	err := col.c.doJSON(http.MethodPost, "/api/collections/"+url.PathEscape(col.name)+"/records", optParam(params), record, &out)
	return out, err
}

// Get retrieves a record by ID.
func (col *Coll[T]) Get(id string, params ...string) (T, error) {
	var out T
	err := col.c.doJSON(http.MethodGet, "/api/collections/"+url.PathEscape(col.name)+"/records/"+url.PathEscape(id), optParam(params), nil, &out)
	return out, err
}

// Update patches a record by ID.
func (col *Coll[T]) Update(id string, patch any, params ...string) (T, error) {
	var out T
	err := col.c.doJSON(http.MethodPatch, "/api/collections/"+url.PathEscape(col.name)+"/records/"+url.PathEscape(id), optParam(params), patch, &out)
	return out, err
}

// Delete removes a record by ID.
func (col *Coll[T]) Delete(id string, params ...string) error {
	return col.c.doJSON(http.MethodDelete, "/api/collections/"+url.PathEscape(col.name)+"/records/"+url.PathEscape(id), optParam(params), nil, nil)
}

// List returns a paginated list response for the collection.
func (col *Coll[T]) List(params ...string) (ListResult[T], error) {
	var out ListResult[T]
	err := col.c.doJSON(http.MethodGet, "/api/collections/"+url.PathEscape(col.name)+"/records", optParam(params), nil, &out)
	return out, err
}

// First returns the first item matching the provided query params.
func (col *Coll[T]) First(params ...string) (T, error) {
	q := optParam(params)
	if q == "" {
		q = "page=1&perPage=1"
	} else {
		q += "&page=1&perPage=1"
	}
	res, err := col.List(q)
	if err != nil {
		var zero T
		return zero, err
	}
	if len(res.Items) == 0 {
		var zero T
		return zero, fmt.Errorf("pbclient: not found")
	}
	return res.Items[0], nil
}
