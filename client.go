package bbs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/gogo/protobuf/proto"
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
	ProtoContentType     = "application/x-protobuf"
)

//go:generate counterfeiter -o fake_bbs/fake_client.go . Client

type Client interface {
	Domains() ([]string, error)
	UpsertDomain(domain string, ttl time.Duration) error
	ActualLRPGroups() (models.ActualLRPGroups, error)
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

func (c *client) ActualLRPGroups() (models.ActualLRPGroups, error) {
	var actualLRPGroups models.ActualLRPGroups
	err := c.doRequest(ActualLRPGroupsRoute, nil, nil, nil, &actualLRPGroups)
	return actualLRPGroups, err
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
		return Error{Type: proto.String(RouterError), Message: &routerError[0]}
	}

	if parsedContentType == JSONContentType {
		return handleJSONResponse(res, responseObject)
	} else if parsedContentType == ProtoContentType {
		protoMessage, ok := responseObject.(proto.Message)
		if !ok {
			return Error{Type: proto.String(InvalidRequest), Message: proto.String("cannot read response body")}
		}
		return handleProtoResponse(res, protoMessage)
	} else {
		return handleNonJSONResponse(res)
	}
}

func handleProtoResponse(res *http.Response, responseObject proto.Message) error {
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Error{Type: proto.String(InvalidResponse), Message: proto.String(err.Error())}
	}

	if res.StatusCode > 299 {
		errResponse := Error{}
		err = proto.Unmarshal(buf, &errResponse)
		if err != nil {
			return Error{Type: proto.String(InvalidProtobufMessage), Message: proto.String(err.Error())}
		}
		return errResponse
	}

	err = proto.Unmarshal(buf, responseObject)
	if err != nil {
		return Error{Type: proto.String(InvalidProtobufMessage), Message: proto.String(err.Error())}
	}
	return nil
}

func handleJSONResponse(res *http.Response, responseObject interface{}) error {
	if res.StatusCode > 299 {
		errResponse := Error{}
		if err := json.NewDecoder(res.Body).Decode(&errResponse); err != nil {
			return Error{Type: proto.String(""), Message: proto.String(err.Error())}
		}
		return errResponse
	}

	if err := json.NewDecoder(res.Body).Decode(responseObject); err != nil {
		return Error{Type: proto.String(""), Message: proto.String(err.Error())}
	}
	return nil
}

func handleNonJSONResponse(res *http.Response) error {
	if res.StatusCode > 299 {
		return Error{
			Type:    proto.String(InvalidResponse),
			Message: proto.String(fmt.Sprintf("Invalid Response with status code: %d", res.StatusCode)),
		}
	}
	return nil
}
