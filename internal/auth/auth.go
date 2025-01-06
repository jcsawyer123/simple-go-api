package auth

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sony/gobreaker"
)

type Client struct {
	baseURL string
	client  *resty.Client
	breaker *gobreaker.CircuitBreaker
	cache   *PermissionCache
}

func NewClient(baseURL string) (*Client, error) {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "auth-service",
		MaxRequests: 3,
		Interval:    0,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("Circuit breaker %s state change: %s -> %s", name, from, to)
		},
	})

	client := resty.New().
		SetTimeout(5 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(100 * time.Millisecond).
		SetRetryMaxWaitTime(2 * time.Second)

	return &Client{
		baseURL: baseURL,
		client:  client,
		breaker: cb,
		cache:   NewPermissionCache(5 * time.Minute),
	}, nil
}

// Implement HealthChecker interface
func (c *Client) Check(ctx context.Context) error {
	_, err := c.client.R().
		SetContext(ctx).
		Get(c.baseURL + "/health")
	return err
}

// Takes token and tries to validate it against /validate (can change to external service)
func (c *Client) ValidateToken(ctx context.Context, token string) error {
	_, err := c.breaker.Execute(func() (interface{}, error) {
		resp, err := c.client.R().
			SetContext(ctx).
			SetHeader("x-aims-auth-token", token).
			// In the future, we will want to use somethning like /aims/v1/:account_id/users/:user_id/roles aswell to get permissions the user has
			Get(c.baseURL + "/aims/v1/token_info")

		if err != nil {
			return nil, fmt.Errorf("auth request failed: %w", err)
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("invalid token: status %d", resp.StatusCode())
		}

		return resp, nil
	})

	return err
}
