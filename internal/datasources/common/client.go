package common

import (
	"net/http"
	"time"
)

type ClientConfig struct {
	BasePath string
	Timeout  time.Duration
}

type Client struct {
	config *ClientConfig
	client *http.Client
}

func NewClient(cfg *ClientConfig) *Client {
	return &Client{
		config: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}
