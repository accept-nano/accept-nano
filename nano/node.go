package nano

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/log"
)

type Node struct {
	url    string
	client http.Client
}

func New(nodeURL string) *Node {
	return &Node{
		url: nodeURL,
	}
}

func (n *Node) SetTimeout(d time.Duration) {
	n.client.Timeout = d
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
	req, err := http.NewRequest("POST", n.url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Debugf("node response: %#v", string(body))
	var errorResponse NodeError
	err = json.Unmarshal(body, &errorResponse)
	if err != nil {
		return err
	}
	if errorResponse.Message != nil {
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
