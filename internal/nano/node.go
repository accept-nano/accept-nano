package nano

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/log"
)

type Node struct {
	url      string
	client   http.Client
	sleep    time.Duration
	auth     string
	apiKey   string
	requestC chan *nodeRequest
	closeC   chan struct{}
}

func New(nodeURL string, timeout, sleep time.Duration, authorization, apiKey string) *Node {
	n := &Node{
		url: nodeURL,
		client: http.Client{
			Timeout: timeout,
		},
		sleep:    sleep,
		auth:     authorization,
		apiKey:   apiKey,
		requestC: make(chan *nodeRequest),
		closeC:   make(chan struct{}),
	}
	go n.caller()
	return n
}

func (n *Node) Close() {
	close(n.closeC)
}

type nodeRequest struct {
	action   string
	args     map[string]interface{}
	response interface{}
	done     chan struct{}
	err      error
}

func (n *Node) caller() {
	for {
		select {
		case req := <-n.requestC:
			req.err = n.callNow(req.action, req.args, req.response)
			close(req.done)
		case <-n.closeC:
			return
		}
		select {
		case <-time.After(n.sleep):
		case <-n.closeC:
			return
		}
	}
}

func (n *Node) call(action string, args map[string]interface{}, response interface{}) error {
	if n.sleep == 0 {
		return n.callNow(action, args, response)
	}
	req := &nodeRequest{action, args, response, make(chan struct{}), nil}
	select {
	case n.requestC <- req:
		<-req.done
		return req.err
	case <-time.After(n.client.Timeout):
		return context.Canceled
	}
}

func (n *Node) callNow(action string, args map[string]interface{}, response interface{}) error {
	if args == nil {
		args = make(map[string]interface{})
	}
	args["action"] = action
	log.Debugf("node request: %#v", args)
	data, err := json.Marshal(args)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, n.url, bytes.NewReader(data)) // nolint:noctx // client timeout set
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if n.auth != "" {
		req.Header.Set("authorization", n.auth)
	}
	if n.apiKey != "" {
		req.Header.Set("api-key", n.apiKey)
	}
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rateLimitRemaining := resp.Header.Get("x-ratelimit-remaining")
	if rateLimitRemaining != "" {
		log.Debugln("Node rate limit remaining:", rateLimitRemaining)
	}
	rateLimitReset := resp.Header.Get("x-ratelimit-reset")
	if rateLimitReset != "" {
		log.Debugln("Node rate limit reset:", rateLimitReset)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Debugf("node response: %d - %#v", resp.StatusCode, string(body))
	var errorResponse NodeError
	err = json.Unmarshal(body, &errorResponse)
	if err == nil && errorResponse.Message != nil {
		return &errorResponse
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
		}
	}
	return json.Unmarshal(body, &response)
}
