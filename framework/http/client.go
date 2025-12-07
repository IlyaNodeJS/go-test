package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"time"
)

// ServiceResolver resolves service names to base URLs.
type ServiceResolver interface {
	Resolve(service string) (string, error)
}

// StaticResolver is a simple resolver backed by a map.
type StaticResolver map[string]string

func (sr StaticResolver) Resolve(service string) (string, error) {
	if base, ok := sr[service]; ok {
		return base, nil
	}
	return fmt.Sprintf("http://%s.local", service), nil
}

// Client orchestrates HTTP calls for declarative tests and suites.
type Client struct {
	client    *nethttp.Client
	resolver  ServiceResolver
	userAgent string
}

func New(resolver ServiceResolver) *Client {
	return &Client{
		client:    &nethttp.Client{Timeout: 30 * time.Second},
		resolver:  resolver,
		userAgent: "go-test-framework/0.1",
	}
}

// Request holds an abstract HTTP request.
type Request struct {
	Service  string
	Method   string
	Endpoint string
	Body     map[string]any
	Headers  map[string]string
}

// Response is a normalized HTTP response.
type Response struct {
	StatusCode int
	Body       map[string]any
	Raw        []byte
}

func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	base, err := c.resolver.Resolve(req.Service)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s%s", base, req.Endpoint)

	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = json.Marshal(req.Body)
		if err != nil {
			return nil, err
		}
	}

	httpReq, err := nethttp.NewRequestWithContext(ctx, req.Method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	if len(bodyBytes) > 0 {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("User-Agent", c.userAgent)
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var body map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			body = map[string]any{"_raw": string(raw)}
		}
	}

	return &Response{StatusCode: resp.StatusCode, Body: body, Raw: raw}, nil
}
