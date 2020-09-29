package nano

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/log"
	"github.com/gorilla/websocket"
)

type Websocket struct {
	url  string
	conn *websocket.Conn
}

func NewWebsocket(wsURL string) *Websocket {
	return &Websocket{
		url: wsURL,
	}
}

func (w *Websocket) Connect(handshakeTimeout time.Duration) error {
	log.Debugf("connecting to websocket: %s", w.url)
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: handshakeTimeout,
	}
	conn, _, err := dialer.Dial(w.url, nil) // nolint:bodyclose
	if err != nil {
		return err
	}
	log.Debugf("connected to websocket: %s", w.url)
	w.conn = conn
	return nil
}

func (w *Websocket) Close() error {
	if w.conn == nil {
		return nil
	}
	return w.conn.Close()
}

type OutgoingMessage struct {
	Action  string                 `json:"action"`
	Topic   string                 `json:"topic,omitempty"`
	Ack     bool                   `json:"ack,omitempty"`
	ID      string                 `json:"id,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

func (m *OutgoingMessage) RequireAck() {
	m.Ack = true
	m.ID = strconv.Itoa(rand.Int()) // nolint: gosec
}

func (w *Websocket) Send(msg OutgoingMessage, timeout time.Duration) error {
	log.Debugf("sending websocket message: %#v", msg)
	_ = w.conn.SetWriteDeadline(deadlineForTimeout(timeout))
	return w.conn.WriteJSON(msg)
}

type IncomingMessage struct {
	Ack     string          `json:"ack"`
	Time    string          `json:"time"`
	ID      string          `json:"id"`
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}

func (w *Websocket) Recv(timeout time.Duration) (msg IncomingMessage, err error) {
	_ = w.conn.SetReadDeadline(deadlineForTimeout(timeout))
	err = w.conn.ReadJSON(&msg)
	if err == nil {
		msg2 := msg
		msg2.Message = nil // prints as hex in log, may be too big for eyes
		log.Debugf("received websocket message: %#v", msg2)
	}
	return
}

func deadlineForTimeout(timeout time.Duration) time.Time {
	if timeout == 0 {
		return time.Time{}
	}
	return time.Now().Add(timeout)
}
