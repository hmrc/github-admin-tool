package graphqlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const GraphqlEndpoint = "https://api.github.com/graphql"

// Client is a client for interacting with a GraphQL API.
type Client struct {
	endpoint   string
	httpClient *http.Client

	// closeReq will close the request body immediately allowing for reuse of client
	closeReq bool
	Log      func(s string)
}

// NewClient makes a new Client capable of making GraphQL requests.
func NewClient() *Client {
	c := &Client{
		endpoint: GraphqlEndpoint,
		Log:      func(string) {},
	}

	c.httpClient = http.DefaultClient

	return c
}

// func (c *Client) logf(format string, args ...interface{}) {
// 	c.Log(fmt.Sprintf(format, args...))
// }

// Run executes the query and unmarshals the response from the data field
// into the response object.
// Pass in a nil response object to skip response parsing.
// If the request fails or the server returns an error, the first error
// will be returned.
func (c *Client) Run(ctx context.Context, req *Request, resp interface{}) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context done: %w", ctx.Err())
	default:
		return c.runWithJSON(ctx, req, resp)
	}
}

func (c *Client) runWithJSON(ctx context.Context, req *Request, resp interface{}) error {
	var requestBody bytes.Buffer

	requestBodyObj := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     req.q,
		Variables: req.vars,
	}

	if err := json.NewEncoder(&requestBody).Encode(requestBodyObj); err != nil {
		return fmt.Errorf("error encode body: %w", err)
	}

	gr := &graphResponse{
		Data: resp,
	}

	r, err := http.NewRequest(http.MethodPost, c.endpoint, &requestBody)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	r.Close = c.closeReq
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	r.Header.Set("Accept", "application/json; charset=utf-8")

	for key, values := range req.Header {
		for _, value := range values {
			r.Header.Add(key, value)
		}
	}

	r = r.WithContext(ctx)

	res, err := c.httpClient.Do(r)
	if err != nil {
		return fmt.Errorf("running do: %w", err)
	}

	defer res.Body.Close()

	var buf bytes.Buffer

	if _, err := io.Copy(&buf, res.Body); err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	if err := json.NewDecoder(&buf).Decode(&gr); err != nil {
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("graphql: server returned a non-200 status code: %w", err)
		}

		return fmt.Errorf("decoding response: %w", err)
	}

	if len(gr.Errors) > 0 {
		return gr.Errors[0]
	}

	return nil
}

// ClientOption are functions that are passed into NewClient to
// modify the behaviour of the Client.
type ClientOption func(*Client)

type graphErr struct {
	Message string
}

func (e graphErr) Error() string {
	return "graphql: " + e.Message
}

type graphResponse struct {
	Data   interface{}
	Errors []graphErr
}

// Request is a GraphQL request.
type Request struct {
	q    string
	vars map[string]interface{}

	// Header represent any request headers that will be set
	// when the request is made.
	Header http.Header
}

// NewRequest makes a new Request with the specified string.
func NewRequest(q string) *Request {
	req := &Request{
		q:      q,
		Header: make(map[string][]string),
	}

	return req
}

// Var sets a variable.
func (req *Request) Var(key string, value interface{}) {
	if req.vars == nil {
		req.vars = make(map[string]interface{})
	}

	req.vars[key] = value
}

// Vars gets the variables for this Request.
func (req *Request) Vars() map[string]interface{} {
	return req.vars
}

// Query gets the query string of this request.
func (req *Request) Query() string {
	return req.q
}
