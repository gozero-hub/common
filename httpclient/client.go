package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type Client struct {
	httpClient  *http.Client
	headers     map[string]string
	retryPolicy RetryPolicy
	logger      logx.Logger
}

// NewClient 创建 HTTP Client（带连接池）
func NewClient(timeout time.Duration, retry RetryPolicy, logger logx.Logger) *Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		// DialContext: (&net.Dialer{
		// 	Timeout: 5 * time.Second,
		// }).DialContext,
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		headers:     make(map[string]string),
		retryPolicy: retry,
		logger:      logger,
	}
}

// SetHeader 设置全局 Header
func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

// Do 通用请求入口（含重试）
func (c *Client) Do(
	ctx context.Context,
	method string,
	rawURL string,
	query map[string]string,
	body interface{},
	headers map[string]string,
) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryPolicy.MaxRetries; attempt++ {
		resp, err := c.doOnce(ctx, method, rawURL, query, body, headers)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		if c.logger != nil {
			c.logger.Error("http request failed", map[string]interface{}{
				"url":     rawURL,
				"method":  method,
				"attempt": attempt,
				"error":   err.Error(),
			})
		}

		time.Sleep(c.retryPolicy.Backoff(attempt))
	}

	return nil, lastErr
}

// 单次请求
func (c *Client) doOnce(
	ctx context.Context,
	method string,
	rawURL string,
	query map[string]string,
	body interface{},
	headers map[string]string,
) (*Response, error) {
	// URL + Query
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if len(query) > 0 {
		q := u.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	// Body
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}

	// 全局 Header
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// 单次 Header
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 是否需要重试
	if c.retryPolicy.RetryCodes[resp.StatusCode] {
		return nil, ErrRetryable
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       data,
		Headers:    resp.Header,
	}, nil
}
