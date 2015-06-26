package bbs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/tedsuo/rata"
)

var ErrReadFromClosedSource = errors.New("read from closed source")
var ErrSendToClosedSource = errors.New("send to closed source")
var ErrSourceAlreadyClosed = errors.New("source already closed")
var ErrSlowConsumer = errors.New("slow consumer")

var ErrSubscribedToClosedHub = errors.New("subscribed to closed hub")
var ErrHubAlreadyClosed = errors.New("hub already closed")

const (
	ContentTypeHeader    = "Content-Type"
	XCfRouterErrorHeader = "X-Cf-Routererror"
	JSONContentType      = "application/json"
)

//go:generate counterfeiter -o fake_bbs/fake_client.go . Client

type Client interface {
	Domains() ([]string, error)
	UpsertDomain(domain string, ttl time.Duration) error
}

func NewClient(url string) Client {
	return &client{
		httpClient:          cf_http.NewClient(),
		streamingHTTPClient: cf_http.NewStreamingClient(),

		reqGen: rata.NewRequestGenerator(url, Routes),
	}
}

type client struct {
	httpClient          *http.Client
	streamingHTTPClient *http.Client

	reqGen *rata.RequestGenerator
}

func (c *client) Domains() ([]string, error) {
	var domains []string
	err := c.doRequest(DomainsRoute, nil, nil, nil, &domains)
	return domains, err
}

func (c *client) UpsertDomain(domain string, ttl time.Duration) error {
	req, err := c.createRequest(UpsertDomainRoute, rata.Params{"domain": domain}, nil, nil)
	if err != nil {
		return err
	}

	if ttl != 0 {
		req.Header.Set("Cache-Control", fmt.Sprintf("max-age=%d", int(ttl.Seconds())))
	}

	return c.do(req, nil)
}

func (c *client) createRequest(requestName string, params rata.Params, queryParams url.Values, request interface{}) (*http.Request, error) {
	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := c.reqGen.CreateRequest(requestName, params, bytes.NewReader(requestJson))
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = queryParams.Encode()
	req.ContentLength = int64(len(requestJson))
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *client) doRequest(requestName string, params rata.Params, queryParams url.Values, request, response interface{}) error {
	req, err := c.createRequest(requestName, params, queryParams, request)
	if err != nil {
		return err
	}
	return c.do(req, response)
}

func (c *client) do(req *http.Request, responseObject interface{}) error {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var parsedContentType string
	if contentType, ok := res.Header[ContentTypeHeader]; ok {
		parsedContentType, _, _ = mime.ParseMediaType(contentType[0])
	}

	if routerError, ok := res.Header[XCfRouterErrorHeader]; ok {
		return Error{Type: RouterError, Message: routerError[0]}
	}

	if parsedContentType == JSONContentType {
		return handleJSONResponse(res, responseObject)
	} else {
		return handleNonJSONResponse(res)
	}
}

func handleJSONResponse(res *http.Response, responseObject interface{}) error {
	if res.StatusCode > 299 {
		errResponse := Error{}
		if err := json.NewDecoder(res.Body).Decode(&errResponse); err != nil {
			return Error{Type: InvalidJSON, Message: err.Error()}
		}
		return errResponse
	}

	if err := json.NewDecoder(res.Body).Decode(responseObject); err != nil {
		return Error{Type: InvalidJSON, Message: err.Error()}
	}
	return nil
}

func handleNonJSONResponse(res *http.Response) error {
	if res.StatusCode > 299 {
		return Error{
			Type:    InvalidResponse,
			Message: fmt.Sprintf("Invalid Response with status code: %d", res.StatusCode),
		}
	}
	return nil
}
