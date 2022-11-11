package nano

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/log"
)

type Node struct {
	url    string
	client http.Client
	auth   string
	apiKey string
}

func New(nodeURL string, timeout time.Duration, authorization, apiKey string) *Node {
	return &Node{
		url: nodeURL,
		client: http.Client{
			Timeout: timeout,
		},
		auth:   authorization,
		apiKey: apiKey,
	}
}

func (n *Node) call(action string, args map[string]interface{}, response interface{}) error {
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
