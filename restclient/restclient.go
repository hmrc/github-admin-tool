package restclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const RestEndpoint = "https://api.github.com"

var errStatusCode = errors.New("returned a non-200 status code")

type Client struct {
	endpoint   string
	token      string
	httpClient *http.Client
	closeReq   bool
}

func NewClient(path, token string) *Client {
	return &Client{
		endpoint:   RestEndpoint + path,
		token:      token,
		httpClient: &http.Client{},
		closeReq:   true,
	}
}

func NewRequest(q string) *Request {
	req := &Request{
		q:      q,
		Header: make(map[string][]string),
	}

	return req
}

type Request struct {
	q string

	// Header represent any request headers that will be set
	// when the request is made.
	Header http.Header
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) Run(ctx context.Context, resp interface{}) (err error) {
	req, err := http.NewRequest(http.MethodGet, c.endpoint, nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Close = c.closeReq

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("running do: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var errRes errorResponse

		if err = json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}

		return fmt.Errorf("incorrect status: %w, %s", errStatusCode, c.endpoint)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("unmarshall: %w", err)
	}

	return nil
}
