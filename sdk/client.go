package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	libraryVersion = "0.1.0"
	defaultBaseURL = "https://api.example.com/"
	userAgent      = "jetson-monitor/" + libraryVersion
	mediaType      = "application/json"

	headerRequestID = "x-request-id"

	internalHeaderRetryAttempts = "X-SDK-Retry-Attempts"

	defaultRetryMax     = 4
	defaultRetryWaitMax = 30
	defaultRetryWaitMin = 1
)

type Client struct {
	HTTPClient *http.Client
	BaseURL    *url.URL
	UserAgent  string

	onRequestCompleted RequestCompletionCallback

	headers map[string]string

	RetryConfig RetryConfig
}

type RetryConfig struct {
	RetryMax     int
	RetryWaitMin *float64
	RetryWaitMax *float64
}

type RequestCompletionCallback func(*http.Request, *http.Response)

type Response struct {
	*http.Response
}

type Links struct{}
type Meta struct{}

type ErrorResponse struct {
	Response  *http.Response
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Attempts  int
}

func (r *ErrorResponse) Error() string {
	var attempted string
	if r.Attempts > 0 {
		attempted = fmt.Sprintf("; giving up after %d attempt(s)", r.Attempts)
	}
	if r.RequestID != "" {
		return fmt.Sprintf("%v %v: %d (request %q) %v%s",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.RequestID, r.Message, attempted)
	}
	return fmt.Sprintf("%v %v: %d %v%s",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message, attempted)
}

type Timestamp struct{ time.Time }

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		HTTPClient: httpClient,
		BaseURL:    baseURL,
		UserAgent:  userAgent,
		headers:    make(map[string]string),
	}

	return c
}

type ClientOpt func(*Client) error

func New(httpClient *http.Client, opts ...ClientOpt) (*Client, error) {
	c := NewClient(httpClient)
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.RetryConfig.RetryMax > 0 {
		retryClient := retryablehttp.NewClient()
		retryClient.RetryMax = c.RetryConfig.RetryMax

		if c.RetryConfig.RetryWaitMin != nil {
			retryClient.RetryWaitMin = time.Duration(*c.RetryConfig.RetryWaitMin * float64(time.Second))
		}
		if c.RetryConfig.RetryWaitMax != nil {
			retryClient.RetryWaitMax = time.Duration(*c.RetryConfig.RetryWaitMax * float64(time.Second))
		}

		retryClient.HTTPClient.Timeout = c.HTTPClient.Timeout

		retryClient.Logger = nil

		retryClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
			if resp != nil {
				resp.Header.Add(internalHeaderRetryAttempts, strconv.Itoa(numTries))
			}
			return resp, err
		}

		retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			if err != nil && strings.Contains(err.Error(), "INTERNAL_ERROR") && strings.Contains(reflect.TypeOf(err).String(), "http2") {
				return true, nil
			}
			return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
		}

		c.HTTPClient = retryClient.StandardClient()
	}

	return c, nil
}

func SetBaseURL(bu string) ClientOpt {
	return func(c *Client) error {
		u, err := url.Parse(bu)
		if err != nil {
			return err
		}
		c.BaseURL = u
		return nil
	}
}

func SetUserAgent(ua string) ClientOpt {
	return func(c *Client) error {
		c.UserAgent = fmt.Sprintf("%s %s", c.UserAgent, ua)
		return nil
	}
}

func SetRequestHeaders(headers map[string]string) ClientOpt {
	return func(c *Client) error {
		for k, v := range headers {
			c.headers[k] = v
		}
		return nil
	}
}

func WithRetryAndBackoffs(retryConfig RetryConfig) ClientOpt {
	return func(c *Client) error {
		c.RetryConfig.RetryMax = retryConfig.RetryMax
		c.RetryConfig.RetryWaitMax = retryConfig.RetryWaitMax
		c.RetryConfig.RetryWaitMin = retryConfig.RetryWaitMin
		return nil
	}
}

func PtrTo[T any](v T) *T { return &v }

func (c *Client) OnRequestCompleted(rc RequestCompletionCallback) {
	c.onRequestCompleted = rc
}

func (c *Client) NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var req *http.Request
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		req, err = http.NewRequest(method, u.String(), nil)
	default:
		buf := new(bytes.Buffer)
		if body != nil {
			if err = json.NewEncoder(buf).Encode(body); err != nil {
				return nil, err
			}
		}
		req, err = http.NewRequest(method, u.String(), buf)
		if err == nil {
			req.Header.Set("Content-Type", mediaType)
		}
	}
	if err != nil {
		return nil, err
	}

	for k, v := range c.headers {
		req.Header.Add(k, v)
	}
	req.Header.Set("Accept", mediaType)
	req.Header.Set("User-Agent", c.UserAgent)
	return req.WithContext(ctx), nil
}

func newResponse(r *http.Response) *Response {
	resp := Response{Response: r}
	return &resp
}

func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {

	resp, err := DoRequestWithClient(ctx, c.HTTPClient, req)

	if err != nil {
		return nil, err
	}

	if c.onRequestCompleted != nil {
		c.onRequestCompleted(req, resp)
	}

	defer func() {
		const maxBodySlurpSize = 2 << 10
		if resp.ContentLength == -1 || resp.ContentLength <= maxBodySlurpSize {
			io.CopyN(io.Discard, resp.Body, maxBodySlurpSize)
		}
		_ = resp.Body.Close()
	}()

	response := newResponse(resp)

	if err := CheckResponse(resp); err != nil {
		return response, err
	}

	if resp.StatusCode != http.StatusNoContent && v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			return response, err
		}
		err = json.NewDecoder(resp.Body).Decode(v)
		return response, err
	}
	return response, nil
}

func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	er := &ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		if uerr := json.Unmarshal(data, er); uerr != nil {
			er.Message = string(data)
		}
	}
	if er.RequestID == "" {
		er.RequestID = r.Header.Get(headerRequestID)
	}
	if attempts, aerr := strconv.Atoi(r.Header.Get(internalHeaderRetryAttempts)); aerr == nil {
		er.Attempts = attempts
	}
	return er
}

func DoRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return DoRequestWithClient(ctx, http.DefaultClient, req)
}

func DoRequestWithClient(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(req.WithContext(ctx))
}
