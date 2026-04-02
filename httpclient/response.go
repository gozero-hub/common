package httpclient

type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
}
