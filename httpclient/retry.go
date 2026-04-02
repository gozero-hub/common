package httpclient

import "time"

type RetryPolicy struct {
	MaxRetries int
	BaseDelay  time.Duration
	RetryCodes map[int]bool
}

// DefaultRetryPolicy 默认重试策略
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  500 * time.Millisecond,
		RetryCodes: map[int]bool{
			500: true,
			502: true,
			503: true,
			504: true,
		},
	}
}

// Backoff 指数退避
func (r RetryPolicy) Backoff(attempt int) time.Duration {
	return time.Duration(1<<attempt) * r.BaseDelay
}
