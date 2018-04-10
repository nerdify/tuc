package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// Client for request balance.
type Client struct {
	Endpoint string
	Token    string
}

// RequestInput is the input for request balance.
type RequestInput struct {
	Card string `json:"card"`
}

// RequestOutput is the output for request balance.
type RequestOutput struct {
	Balance    float64 `json:"balance,string"`
	Message    string  `json:"message"`
	StatusCode int     `json:"status_code"`
}

// NewClient create a new client.
func NewClient(url, token string) *Client {
	return &Client{
		Endpoint: url,
		Token:    token,
	}
}

// GetBalance get the balance for the given card.
func (c *Client) GetBalance(in *RequestInput) (*RequestOutput, error) {
	b, err := json.Marshal(in)

	if err != nil {
		return nil, errors.Wrap(err, "marshaling input")
	}

	req, err := http.NewRequest(http.MethodPost, c.Endpoint, bytes.NewReader(b))

	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "requesting")
	}

	defer res.Body.Close()

	out := new(RequestOutput)

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return nil, errors.Wrap(err, "unmarshaling output")
	}

	return out, nil
}
