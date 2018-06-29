package client

import (
	"encoding/json"
	"fmt"
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
	Code int `json:"code"`
	Data []struct {
		Balance float64 `json:"balance,string"`
		Status  string  `json:"status"`
	} `json:"data"`
}

// NewClient create a new client.
func NewClient(url string) *Client {
	return &Client{
		Endpoint: url,
	}
}

// GetBalance get the balance for the given card.
func (c *Client) GetBalance(in *RequestInput) (*RequestOutput, error) {
	url := fmt.Sprintf("%s/%s", c.Endpoint, in.Card)
	res, err := http.Get(url)

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
