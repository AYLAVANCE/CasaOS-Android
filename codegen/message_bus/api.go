// Package message_bus provides types for CasaOS MessageBus integration.
package message_bus

import (
	"context"
	"net/http"
)

// EventType defines the structure for MessageBus event types.
type EventType struct {
	Name             string         `json:"name"`
	SourceID         string         `json:"sourceID"`
	PropertyTypeList []PropertyType `json:"propertyTypeList"`
}

// PropertyType defines the structure for event property types.
type PropertyType struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Example     *string `json:"example,omitempty"`
}

// Event represents an event published to the MessageBus.
type Event struct {
	SourceID   string     `json:"sourceID"`
	Name       string     `json:"name"`
	Properties []Property `json:"properties"`
	Timestamp  int64      `json:"timestamp"`
	UUID       string     `json:"uuid"`
}

// Property is a key-value pair attached to an event.
type Property struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Client is the base HTTP client for MessageBus communication.
type Client struct {
	Server string
}

// ClientOption is a function that modifies a Client.
type ClientOption func(*Client) error

// ClientWithResponses is a MessageBus client that parses responses.
type ClientWithResponses struct {
	Client
}

// NewClientWithResponses creates a new MessageBus client.
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	c := &ClientWithResponses{
		Client: Client{Server: server},
	}
	for _, opt := range opts {
		if err := opt(&c.Client); err != nil {
			return c, err
		}
	}
	return c, nil
}

// RegisterEventTypesResponse wraps the HTTP response from a register event types call.
type RegisterEventTypesResponse struct {
	HTTPResponse *http.Response
	Body         []byte
}

// StatusCode returns the HTTP status code of the response.
func (r *RegisterEventTypesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 200
}

// Status returns the HTTP status string.
func (r *RegisterEventTypesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return "200 OK"
}

// RegisterEventTypesWithResponse registers event types with the MessageBus.
func (c *ClientWithResponses) RegisterEventTypesWithResponse(ctx context.Context, eventTypes []EventType) (*RegisterEventTypesResponse, error) {
	// Stub: actual implementation communicates with CasaOS-MessageBus service
	return &RegisterEventTypesResponse{}, nil
}

// PublishEventResponse wraps the HTTP response from a publish event call.
type PublishEventResponse struct {
	HTTPResponse *http.Response
}

// StatusCode returns the HTTP status code of the response.
func (r *PublishEventResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 200
}

// Status returns the HTTP status string.
func (r *PublishEventResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return "200 OK"
}

// PublishEventWithResponse publishes an event to the MessageBus.
func (c *ClientWithResponses) PublishEventWithResponse(ctx context.Context, sourceID string, name string, properties map[string]string) (*PublishEventResponse, error) {
	// Stub: actual implementation sends event via HTTP to CasaOS-MessageBus service
	return &PublishEventResponse{}, nil
}
