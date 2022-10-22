package subscriber

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/accept-nano/accept-nano/internal/nano"
	"github.com/cenkalti/log"
)

type Subscriber struct {
	Confirmations       chan string
	url                 string
	handshakeTimeout    time.Duration
	writeTimeout        time.Duration
	ackTimeout          time.Duration
	keepaliveDuration   time.Duration
	newConnectionC      chan *nano.Websocket
	updateSubscriptionC chan updateSubscription
	closeC              chan struct{}

	m           sync.Mutex
	pendingAcks map[string]chan struct{}
}

type updateSubscription struct {
	Action  string
	Account string
}

type confirmationMessage struct {
	FromAccount string `json:"account"`
	Block       struct {
		ToAccount string `json:"link_as_account"`
	} `json:"block"`
}

func New(wsURL string, handshakeTimeout, writeTimeout, ackTimeout, keepaliveDuration time.Duration) *Subscriber {
	return &Subscriber{
		Confirmations:       make(chan string),
		url:                 wsURL,
		handshakeTimeout:    handshakeTimeout,
		writeTimeout:        writeTimeout,
		ackTimeout:          ackTimeout,
		keepaliveDuration:   keepaliveDuration,
		newConnectionC:      make(chan *nano.Websocket),
		updateSubscriptionC: make(chan updateSubscription),
		closeC:              make(chan struct{}),
		pendingAcks:         make(map[string]chan struct{}),
	}
}

func (s *Subscriber) Close() {
	close(s.closeC)
}

func (s *Subscriber) Run() {
	go s.writer()
	for {
		s.reader()
		select {
		case <-s.closeC:
			return
		default:
			time.Sleep(time.Second)
		}
	}
}

func (s *Subscriber) writer() {
	var ws *nano.Websocket
	connected := func() bool { return ws != nil }

	var subscribed, toAdd, toDel map[string]struct{}
	subscribed = make(map[string]struct{})

	const d = time.Second
	t := NewTimer(d)
	defer t.Stop()

	keepAlive := time.NewTimer(s.keepaliveDuration)
	defer t.Stop()

	handleError := func(err error) {
		log.Errorln("websocket send error:", err.Error())
		ws.Close()
		ws = nil
	}

	for {
		select {
		case ws = <-s.newConnectionC:
			toAdd = make(map[string]struct{})
			toDel = make(map[string]struct{})
			t.Stop()
			msg := nano.OutgoingMessage{
				Action: "subscribe",
				Topic:  "confirmation",
				Options: map[string]interface{}{
					"include_election_info": "false",
					"include_block":         "true",
					"accounts":              stringList(subscribed),
				},
			}
			err := s.sendWithTimeout(ws, msg)
			if err != nil {
				handleError(err)
				break
			}
		case action := <-s.updateSubscriptionC:
			switch action.Action {
			case "add":
				subscribed[action.Account] = struct{}{}
				if connected() {
					toAdd[action.Account] = struct{}{}
					delete(toDel, action.Account)
					t.Delay()
				}
			case "del":
				delete(subscribed, action.Account)
				if connected() {
					toDel[action.Action] = struct{}{}
					delete(toAdd, action.Action)
					t.Delay()
				}
			}
		case <-t.C:
			t.SetNotRunning()
			if !connected() {
				continue
			}
			msg := nano.OutgoingMessage{
				Action: "update",
				Topic:  "confirmation",
				Options: map[string]interface{}{
					"accounts_add": stringList(toAdd),
					"accounts_del": stringList(toDel),
				},
			}
			err := s.sendWithTimeout(ws, msg)
			if err != nil {
				handleError(err)
				break
			}
		case <-keepAlive.C:
			if !connected() {
				continue
			}
			msg := nano.OutgoingMessage{
				Action: "ping",
			}
			err := s.sendWithTimeout(ws, msg)
			if err != nil {
				handleError(err)
				break
			}
		case <-s.closeC:
			if ws != nil {
				ws.Close()
			}
			return
		}
	}
}

func (s *Subscriber) reader() {
	ws := nano.NewWebsocket(s.url)
	err := ws.Connect(s.handshakeTimeout)
	if err != nil {
		log.Errorln("websocket connect error:", err.Error())
		return
	}

	// Make sure connection is closed on return
	closed := make(chan struct{})
	defer func() {
		ws.Close()
		close(closed)
	}()
	go func() {
		select {
		case <-closed:
		case <-s.closeC:
			ws.Close()
		}
	}()

	// Notify writer for new connection
	select {
	case s.newConnectionC <- ws:
	case <-s.closeC:
		return
	}

	for {
		msg, err := ws.Recv(0)
		if err != nil {
			log.Errorln("websocket receive error:", err.Error())
			return
		}
		switch {
		case msg.Ack != "":
			s.m.Lock()
			if ch, ok := s.pendingAcks[msg.ID]; ok {
				close(ch)
				delete(s.pendingAcks, msg.ID)
			}
			s.m.Unlock()
		case msg.Topic == "confirmation":
			var cf confirmationMessage
			err = json.Unmarshal(msg.Message, &cf)
			if err != nil {
				log.Errorln("cannot unmarshal confirmation message:", err.Error())
				break
			}
			select {
			case s.Confirmations <- cf.Block.ToAccount:
			case <-s.closeC:
				return
			}
		}
	}
}

func (s *Subscriber) Subscribe(account string) {
	log.Debugln("subscribing confirmations for account:", account)
	select {
	case s.updateSubscriptionC <- updateSubscription{Action: "add", Account: account}:
	case <-s.closeC:
	}
}

func (s *Subscriber) Unsubscribe(account string) {
	log.Debugln("unsubscribing confirmations for account:", account)
	select {
	case s.updateSubscriptionC <- updateSubscription{Action: "del", Account: account}:
	case <-s.closeC:
	}
}

func (s *Subscriber) sendWithTimeout(ws *nano.Websocket, msg nano.OutgoingMessage) error {
	msg.RequireAck()
	ch := make(chan struct{})
	s.m.Lock()
	s.pendingAcks[msg.ID] = ch
	s.m.Unlock()
	err := ws.Send(msg, s.writeTimeout)
	if err != nil {
		ws.Close()
		return err
	}
	go func() {
		select {
		case <-time.After(s.ackTimeout):
			log.Errorln("websocket request timed out: %s", msg.Action)
			ws.Close()
		case <-ch:
		case <-s.closeC:
		}
	}()
	return nil
}

func stringList(m map[string]struct{}) []string {
	l := make([]string, 0, len(m))
	for key := range m {
		l = append(l, key)
	}
	return l
}
